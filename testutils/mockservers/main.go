package main

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/valyala/fasthttp"
)

type seconds = time.Duration

func sleepFor(name string, n seconds) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		log.Printf("%q server got request", name)
		time.Sleep(n * time.Second)
	}
}

func serveError(ctx *fasthttp.RequestCtx) {
	log.Println("fake error server got request")
	ctx.Error("fake", 500)
}

func main() {
	log.SetOutput(ioutil.Discard)
	s1 := &fasthttp.Server{
		Handler: sleepFor("no sleep", 0),
	}
	s2 := &fasthttp.Server{
		Handler: sleepFor("sleep for 1 sec", 1),
	}
	s3 := &fasthttp.Server{
		Handler: serveError,
	}

	go func() { must(s1.ListenAndServe(":9111")) }()
	go func() { must(s2.ListenAndServe(":9112")) }()
	must(s3.ListenAndServe(":9113"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
