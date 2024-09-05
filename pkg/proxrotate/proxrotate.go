package proxrotate

import (
	"net/http"
	"net/url"
	"sync"
)

type ProxyRotator struct {
	proxs []*url.URL
	rwmx  sync.RWMutex
	next  int
}

func NewProxyRotator(proxs []*url.URL) *ProxyRotator {
	proxrot := ProxyRotator{
		proxs: proxs,
	}

	return &proxrot
}

func (proxrot *ProxyRotator) Rotate(cl *http.Client) *http.Client {
	if cl == nil {
		cl = http.DefaultClient
	}

	return proxrot.rotate(cl)
}

func (proxrot *ProxyRotator) rotate(cl *http.Client) *http.Client {
	proxrot.rwmx.RLock()
	defer proxrot.rwmx.RUnlock()

	var prox *url.URL
	if proxrot.next > len(proxrot.proxs)-1 {
		proxrot.next = 0
	}

	if len(proxrot.proxs) != 0 {
		prox = proxrot.proxs[proxrot.next]
	}

	if prox != nil {
		cl.Transport = &http.Transport{Proxy: http.ProxyURL(prox)}
	}

	proxrot.next++

	return cl
}
