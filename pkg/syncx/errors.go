package syncx

import "sync"

// Errors is a simple concurrent-safe error collector.
type Errors struct {
	errs []error
	mx   sync.Mutex
}

func (es *Errors) Add(err error) {
	es.mx.Lock()
	defer es.mx.Unlock()
	es.errs = append(es.errs, err)
}

func (es *Errors) Errors() []error {
	es.mx.Lock()
	defer es.mx.Unlock()
	return es.errs
}
