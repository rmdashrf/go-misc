package tlsmisc

import (
	"crypto/tls"
	"sync"
	"sync/atomic"
	"time"

	fsnotify "gopkg.in/fsnotify.v1"
)

type CertificateGetter interface {
	GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error)
}

type ReloadingCertificateGetter struct {
	// type: *tls.Certificate
	currentCert atomic.Value
	reloadCh    chan struct{}
	stopCh      chan struct{}

	errorCbMu sync.Mutex
	errorCb   func(err error)

	certfile string
	keyfile  string

	opts *options
}

type ReloadError struct {
	Certfile string
	Keyfile  string
	Err      error
}

func (r ReloadError) Error() string {
	return r.Err.Error()
}

type WatcherError struct {
	error
}

/// Returns the currently loaded certificate from memory.
func (r *ReloadingCertificateGetter) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return r.currentCert.Load().(*tls.Certificate), nil
}

func (r *ReloadingCertificateGetter) Stop() {
	close(r.stopCh)
}

/// Registers a callback to be called whenever an error is encountered
/// Possible errors that can be returned are *ReloadError, *WatchError
func (r *ReloadingCertificateGetter) ErrorCallback(f func(err error)) {
	r.errorCbMu.Lock()
	r.errorCb = f
	r.errorCbMu.Unlock()
}

/// Reloads the certificate from disk. Any errors encountered will be reported
/// through the registered error callback.
func (r *ReloadingCertificateGetter) Reload() {
	cert, err := tls.LoadX509KeyPair(r.certfile, r.keyfile)
	if err != nil {
		// Error while reloading. Keep the current cert unmodified, but trigger
		// an error callback
		r.fireError(&ReloadError{
			Err:      err,
			Certfile: r.certfile,
			Keyfile:  r.keyfile,
		})
		return
	}

	// Certificate load successful
	r.currentCert.Store(&cert)
}

func (r *ReloadingCertificateGetter) fireError(err error) {
	r.errorCbMu.Lock()
	cb := r.errorCb
	r.errorCbMu.Unlock()

	if cb == nil {
		return
	}

	cb(err)
}

func (r *ReloadingCertificateGetter) fileWatch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		r.fireError(&WatcherError{err})
		return
	}

	if err := watcher.Add(r.certfile); err != nil {
		r.fireError(&WatcherError{err})
		return
	}

	if err := watcher.Add(r.keyfile); err != nil {
		r.fireError(&WatcherError{err})
		return
	}

	t := time.NewTimer(0)
	<-t.C
LOOP:
	for {
		select {
		case <-r.stopCh:
			break LOOP
		case err := <-watcher.Errors:
			{
				r.fireError(&WatcherError{err})
				continue
			}
		case event := <-watcher.Events:
			{
				if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
					if r.opts.CoalescingTimeout == 0 {
						t.Reset(100 * time.Millisecond)
					} else {
						t.Reset(r.opts.CoalescingTimeout)
					}
				}
			}

		case <-t.C:
			{
				// Aggregation timer has fired, reload the certificate.
				r.Reload()
			}
		}
	}

	t.Stop()
	watcher.Close()

}

func (r *ReloadingCertificateGetter) run() {
	if !r.opts.DisableFilesystemWatch {
		go r.fileWatch()
	}

	for {
		select {
		case <-r.stopCh:
			return
		case <-r.reloadCh:
			{
				r.Reload()
			}
		}
	}
}

type options struct {
	DisableFilesystemWatch bool
	CoalescingTimeout      time.Duration
}

type Option func(*options)

func DisableFsWatch(opt *options) {
	opt.DisableFilesystemWatch = true
}

func WithCoalescingTimeout(d time.Duration) Option {
	return func(opt *options) {
		opt.CoalescingTimeout = d
	}
}

func ReloadingCertificateGetterFromFile(certfile, keyfile string, opts ...Option) (*ReloadingCertificateGetter, error) {
	cert, err := tls.LoadX509KeyPair(certfile, keyfile)
	if err != nil {
		return nil, err
	}

	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	cg := &ReloadingCertificateGetter{
		reloadCh: make(chan struct{}, 1),
		certfile: certfile,
		keyfile:  keyfile,
		opts:     options,
	}

	cg.currentCert.Store(&cert)
	go cg.run()

	return cg, nil
}

func ReloadingCertificateTlsConfig(certfile, keyfile string, opts ...Option) (*tls.Config, error) {
	cg, err := ReloadingCertificateGetterFromFile(certfile, keyfile, opts...)
	if err != nil {
		return nil, err
	}

	tlsConf := &tls.Config{
		GetCertificate: cg.GetCertificate,
	}

	return tlsConf, nil
}
