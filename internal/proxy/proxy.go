package proxy

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

type Proxy struct {
	targetURL *atomic.Value
	transport *http.Transport
}

func New() *Proxy {
	p := &Proxy{
		targetURL: &atomic.Value{},
		transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	p.targetURL.Store((*url.URL)(nil))
	return p
}

func (p *Proxy) SetTarget(socketPath string) {
	if socketPath == "" {
		p.targetURL.Store((*url.URL)(nil))
		return
	}

	p.transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
		return net.Dial("unix", socketPath)
	}

	targetURL, _ := url.Parse("http://localhost")
	p.targetURL.Store(targetURL)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target := p.targetURL.Load().(*url.URL)
	if target == nil {
		http.Error(w, "no target available", http.StatusServiceUnavailable)
		return
	}

	httputil.NewSingleHostReverseProxy(target).ServeHTTP(w, r)
}
