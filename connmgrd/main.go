package main

import (
	"log"
	"time"

	"github.com/highloadltd/connmgr"
	"github.com/valyala/fasthttp"
)

type state struct {
	cl       *fasthttp.Client
	httpPool *connmgr.Pool
}

func (s *state) proxy(ctx *fasthttp.RequestCtx) {
	h := s.httpPool.GetNextHost()
	if h == nil {
		ctx.Error("Service Temporary Unavailable", 502)
		return
	}
	addr := h.Addr()

	ctx.Request.SetRequestURI(addr)
	err := s.cl.Do(&ctx.Request, &ctx.Response)
	if err != nil {
		log.Printf("error while processing request: %s", err)
	}
}

func main() {
	s := &state{
		httpPool: connmgr.NewPool(),
		cl: &fasthttp.Client{
			MaxConnsPerHost:               fasthttp.DefaultMaxConnsPerHost * 300,
			DisableHeaderNamesNormalizing: true,
		},
	}
	h1 := connmgr.NewHost(connmgr.HostTypeHTTP, "http://localhost:9111", "http://localhost:9111")
	h1.SetTimeout(100 * time.Millisecond)
	h2 := connmgr.NewHost(connmgr.HostTypeHTTP, "http://localhost:9112", "http://localhost:9112")
	h2.SetTimeout(100 * time.Millisecond)
	s.httpPool.Add(h1)
	s.httpPool.Add(h2)
	server := fasthttp.Server{
		Handler: s.proxy,
	}
	must(server.ListenAndServe(":9110"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
