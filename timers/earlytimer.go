package timers

import (
	"sync"
	"time"
)

// EarlyPeriodicTimer is a timer will periodically invoke a given task. However,
// it also has the option to start the task ahead of time. When a task has been
// prematurely started, the timer will reset.
type EarlyPeriodicTimer struct {
	timerMu sync.Mutex
	timer   *time.Timer
	period  time.Duration

	task func()

	stopChan chan struct{}
}

type earlyTimerOpts struct {
	stopChan   chan struct{}
	runOnStart bool
}

type EarlyTimerOption func(*earlyTimerOpts)

func WithStopCh(stopCh chan struct{}) EarlyTimerOption {
	return func(opts *earlyTimerOpts) {
		opts.stopChan = stopCh
	}
}

func RunOnStart() EarlyTimerOption {
	return func(opts *earlyTimerOpts) {
		opts.runOnStart = true
	}
}

func NewEarlyPeriodicTimer(period time.Duration, task func(), opts ...EarlyTimerOption) *EarlyPeriodicTimer {
	if task == nil {
		panic("Empty task")
	}

	var options earlyTimerOpts
	for _, opt := range opts {
		opt(&options)
	}

	if options.stopChan == nil {
		options.stopChan = make(chan struct{})
	}

	timer := time.NewTimer(0)
	timer.Stop()

	ret := &EarlyPeriodicTimer{
		timer:    timer,
		task:     task,
		period:   period,
		stopChan: options.stopChan,
	}

	go ret.run(options.runOnStart)
	return ret
}

func (p *EarlyPeriodicTimer) startTimer() {
	p.timerMu.Lock()
	p.timer.Reset(p.period)
	p.timerMu.Unlock()
}

func (p *EarlyPeriodicTimer) stopTimer() {
	p.timerMu.Lock()
	p.timer.Stop()
	p.timerMu.Unlock()
}

func (p *EarlyPeriodicTimer) Stop() {
	close(p.stopChan)
}

func (p *EarlyPeriodicTimer) RunNow() {
	p.stopTimer()
	p.task()
	p.startTimer()
}

func (p *EarlyPeriodicTimer) run(runOnStart bool) {
	if runOnStart {
		p.task()
	}

	p.startTimer()
	for {
		select {
		case <-p.timer.C:
			{
				p.RunNow()
			}
		case <-p.stopChan:
			{
				p.stopTimer()
				return
			}
		}
	}
}
