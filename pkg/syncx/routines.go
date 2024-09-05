package syncx

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrDied          = errors.New("routines already died")
	ErrLimitExceeded = errors.New("goroutine count limit is exceeded")
)

// Routines manages the lifecycle of goroutines.
// It provides a simple functionality to manage your goroutine
// operations, from running to gracefuly kill the goroutines.
type Routines struct {
	ctx        context.Context
	startC     chan bool
	dyingC     chan bool
	wg         WaitGroup
	errs       Errors
	mx         sync.Mutex
	killReason string
	limit      int
}

func NewRoutines() *Routines {
	return &Routines{
		startC: make(chan bool),
		dyingC: make(chan bool),
		wg:     WaitGroup{},
		errs:   Errors{},
		mx:     sync.Mutex{},
	}
}

// WithCtx tells Routines to listen to cancellation signal 
// from [context.Context]. Upon receiving cancellation signal,
// Routines will close dyingC channel to kill all goroutines.
// The context that is used must have an active cancellation
// signal.
func (rc *Routines) WithCtx(ctx context.Context) {
	rc.ctx = ctx
}

// WithLimit limits how much goroutines that can be run
// at the same time.
func (rc *Routines) WithLimit(lim int) {
	rc.limit = lim
}

// Go register a goroutine to be run after Routines.Run
// is called
func (rc *Routines) Go(f func() error) error {
	if rc.isDying() {
		return ErrDied
	}

	if rc.limit > 0 && rc.wg.Count() >= int64(rc.limit) {
		return ErrLimitExceeded
	}

	rc.wg.Add(1)
	go rc._go(f)

	return nil
}

func (rc *Routines) _go(f func() error) {
	defer rc.wg.Done()

	<-rc.startC

	if err := f(); err != nil {
		rc.errs.Add(err)
	}
}

// Dying return a channel to listen to dying channel.
// The returned channel is used as an indicator for
// when client trying to kill all of the goroutines.
// Use this channel to shutdown the goroutines gracefuly
func (rc *Routines) Dying() <-chan bool {
	return rc.dyingC
}

// Run run all the registered goroutines
func (rc *Routines) Run() error {
	if rc.isDying() {
		return ErrDied
	}

	go rc._waitCtxCancel()

	close(rc.startC)
	return nil
}

// Kill kills the goroutines by closing rc.dyingC.
// The closing of rc.dyingC will be listened by the 
// goroutines as a signal to stop or shutdown. We
// literally telling the goroutines to kill itself
func (rc *Routines) Kill(reason string) error {
	rc.mx.Lock()
	defer rc.mx.Unlock()

	if rc.isDying() {
		return ErrDied
	}

	rc.killReason = reason

	close(rc.dyingC)

	return nil
}
// KillReason get the reason why the goroutines
// is killed
func (rc *Routines) KillReason() (string, bool) {
	if rc.isDying() {
		rc.mx.Lock()
		defer rc.mx.Unlock()
		return rc.killReason, true
	}

	return "", false
}

// Wait waits until all of the goroutines returns or killed
func (rc *Routines) Wait() {
	rc.wg.Wait()
}

// WaitAvailable will block until Routines is available to run
// another goroutine. If rc.limit <= 0, call to this method
// will never block
func (rc *Routines) WaitAvailable() {
	if rc.isDying() {
		return
	}

	if rc.limit <= 0 {
		return
	}

	for rc.wg.Count() >= int64(rc.limit) {
		// waiting until available to
		// run another goroutine
	}
}

func (rc *Routines) _waitCtxCancel() {
	if rc.ctx != nil {
		c := rc.ctx.Done()
		if c != nil {
			select {
			case <-c:
				close(rc.dyingC)
			case <-rc.dyingC:
				return
			}
		}
	}
}

func (rc *Routines) Errors() []error {
	return rc.errs.Errors()
}

func (rc *Routines) isDying() bool {
	select {
	case <-rc.dyingC:
		return true
	default:
	}

	return false
}
