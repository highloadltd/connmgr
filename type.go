package connmgr

import (
	"errors"
	"math"
	"time"
)

type HostType uint8
type Resolver interface {
	Do(addr string) error
	Load(cfg KV) error
}

const (
	HostTypeHTTP HostType = iota
)

var ErrNoResolver = errors.New("no resolver for custom type")

const maxDuration = time.Duration(math.MaxInt64)

// Host handles an abstract host.
// It could be raw TCP/UDP, or something like HTTP,
// for that reason Addr is a string.
type Host struct {
	htype    HostType
	pingAddr string
	addr     string
	config   KV
}

func (h *Host) Type() HostType {
	return h.htype
}

func (h *Host) Addr() string {
	return h.addr
}

func (h *Host) PingAddr() string {
	return h.pingAddr
}

func (h *Host) Config() KV {
	return h.config
}

// NewHost creates a new Host instance
func NewHost(t HostType, addr, pingAddr string) *Host {
	return &Host{
		htype:    t,
		pingAddr: pingAddr,
		addr:     addr,
	}
}

func (h *Host) SetTimeout(t time.Duration) {
	h.config.Set("timeout", t)
}

func (h *Host) RTT() (time.Duration, error) {
	var ok bool
	var resolver Resolver
	if resolver, ok = h.getResolver(); !ok {
		return maxDuration, ErrNoResolver
	}
	if err := resolver.Load(h.config); err != nil {
		return maxDuration, err
	}
	start := nanotime()
	if err := resolver.Do(h.pingAddr); err != nil {
		return maxDuration, err
	}
	end := nanotime()
	return time.Duration(time.Duration(end-start) * time.Nanosecond), nil
}

func (h *Host) getResolver() (r Resolver, ok bool) {
	createResolver, ok := getResolverFor(h.htype)
	if !ok {
		return nil, false
	}
	return createResolver(), true
}
