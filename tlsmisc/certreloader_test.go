package tlsmisc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type cgMock struct{}

func (c cgMock) GetCertificate(ch *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return nil, nil
}

func TestCompiles(t *testing.T) {
	var (
		cg   CertificateGetter
		mock cgMock
	)

	cg = &mock

	config := tls.Config{
		GetCertificate: cg.GetCertificate,
	}

	_ = config
}

func withTestDirEnv(t *testing.T, base string, f func(dir string)) {
	dir, err := ioutil.TempDir("", "rmdashrf-tlsmisc-test-")
	if err != nil {
		t.Fatal("Could not create temporary directory", err)
	}

	// Copy over base.pem and base-key.pem
	cp(
		filepath.Join("testdata", fmt.Sprintf("%s.pem", base)),
		filepath.Join(dir, fmt.Sprintf("%s.pem", base)),
	)

	cp(
		filepath.Join("testdata", fmt.Sprintf("%s-key.pem", base)),
		filepath.Join(dir, fmt.Sprintf("%s-key.pem", base)),
	)

	f(dir)
	os.RemoveAll(dir)

}

func cp(from, to string) error {
	f, err := os.Open(from)
	defer f.Close()
	if err != nil {
		return err
	}

	g, err := os.Create(to)
	defer g.Close()
	if err != nil {
		return err
	}

	if _, err := io.Copy(g, f); err != nil {
		return err
	}

	return nil
}

func assertCn(t *testing.T, cg CertificateGetter, cn string) {
	cert, err := cg.GetCertificate(nil)
	if err != nil {
		t.Fatal("Failed to GetCertificate", err)
	}

	c, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatal("Error while parsing x509 data", err)
	}

	if c.Subject.CommonName != cn {
		t.Fatalf("CN mismatch. Expected %s, got %s\n", cn, c.Subject.CommonName)
	}
}

func TestManualReload(t *testing.T) {
	withTestDirEnv(t, "test1", func(dir string) {
		cg, err := ReloadingCertificateGetterFromFile(
			filepath.Join(dir, "test1.pem"),
			filepath.Join(dir, "test1-key.pem"),
			DisableFsWatch,
		)

		cg.ErrorCallback(func(err error) {
			t.Fatal("Got unexpected error callback", err)
		})

		if err != nil {
			t.Fatal("Could not create certificate getter", err)
		}

		assertCn(t, cg, "test certificate 1")
		cg.Reload()

		assertCn(t, cg, "test certificate 1")

		// Now copy over test2
		cp(
			filepath.Join("testdata", "test2.pem"),
			filepath.Join(dir, "test1.pem"),
		)

		cp(
			filepath.Join("testdata", "test2-key.pem"),
			filepath.Join(dir, "test1-key.pem"),
		)

		assertCn(t, cg, "test certificate 1")
		cg.Reload()
		assertCn(t, cg, "test certificate 2")

	})
}

func TestFilesystemReload(t *testing.T) {
	withTestDirEnv(t, "test1", func(dir string) {
		cg, err := ReloadingCertificateGetterFromFile(
			filepath.Join(dir, "test1.pem"),
			filepath.Join(dir, "test1-key.pem"),
		)

		cg.ErrorCallback(func(err error) {
			t.Fatal("Got unexpected error callback", err)
		})

		if err != nil {
			t.Fatal("Could not create certificate getter", err)
		}

		time.Sleep(200 * time.Millisecond)

		assertCn(t, cg, "test certificate 1")
		// Now copy over test2
		cp(
			filepath.Join("testdata", "test2.pem"),
			filepath.Join(dir, "test1.pem"),
		)

		cp(
			filepath.Join("testdata", "test2-key.pem"),
			filepath.Join(dir, "test1-key.pem"),
		)

		time.Sleep(200 * time.Millisecond)

		assertCn(t, cg, "test certificate 2")
	})
}
