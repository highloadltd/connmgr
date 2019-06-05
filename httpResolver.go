package connmgr

import (
	"fmt"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

type HTTPResolver struct {
	method  string
	body    string
	timeout time.Duration
}

func NewHTTPResolver() Resolver {
	return &HTTPResolver{}
}

func (r *HTTPResolver) Load(cfg KV) error {
	m := cfg.GetWithDefault("method", "GET")
	method, ok := m.(string)
	if !ok {
		return fmt.Errorf("invalid method type: got %T", m)
	}

	r.method = strings.ToUpper(method)

	b := cfg.GetWithDefault("body", "")
	body, ok := b.(string)
	if ok && body != "" && (r.method == "GET" || r.method == "HEAD") {
		return fmt.Errorf("unexpected body for method %q", r.method)
	}

	r.body = body

	t := cfg.GetWithDefault("timeout", time.Duration(0))
	timeout, ok := t.(time.Duration)
	if !ok {
		return fmt.Errorf("unexpected timeout type: %T", t)
	}
	if timeout >= 0 {
		r.timeout = timeout
	}

	return nil
}

func (r *HTTPResolver) Do(addr string) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	req.Header.SetMethod(r.method)
	req.SetBodyString(r.body)
	req.SetRequestURI(addr)
	var err error
	if r.timeout != 0 {
		err = fasthttp.DoDeadline(req, resp, time.Now().Add(r.timeout))
	} else {
		err = fasthttp.Do(req, resp)
	}

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)

	return err
}
