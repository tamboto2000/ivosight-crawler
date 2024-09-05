package syncx

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrNegativeCounter = errors.New("negative counter")

// WaitGroup is actually a thin wrapper around [sync.WaitGroup],
// but with added functionality.
// Unlike [sync.WaitGroup], it has additional method to count
// the total of goroutines that is running.
type WaitGroup struct {
	wg sync.WaitGroup
	c  atomic.Int64
}

// Add is a thin wrapper around [sync.WaitGroup.Add]
func (w *WaitGroup) Add(d int) {
	w.c.Add(int64(d))
	w.wg.Add(d)
}

// Done is a thin wrapper around [sync.WaitGroup.Done].
// It will return error ErrNegativeCounter if the internal
// counter already at 0
func (w *WaitGroup) Done() error {
	if w.c.Load() <= 0 {
		return ErrNegativeCounter
	}

	w.c.Add(int64(-1))
	w.wg.Done()

	return nil
}

// Wait is a thin wrapper around [sync.WaitGroup.Wait]
func (w *WaitGroup) Wait() {
	w.wg.Wait()
}

// Count get the count of currently running goroutines.
func (w *WaitGroup) Count() int64 {
	return w.c.Load()
}
