package connmgr

import (
	"sync"
)

type resolvers struct {
	reg map[HostType]func() Resolver
	mu  sync.RWMutex
}

var resolverRegistry = &resolvers{
	mu: sync.RWMutex{},
	reg: map[HostType]func() Resolver{
		HostTypeHTTP: NewHTTPResolver,
	},
}

// AddResolverFor given HostType. Resolver's methods must not panic (we do not handle panics).
func AddResolverFor(t HostType, createResolver func() Resolver) {
	resolverRegistry.mu.Lock()
	resolverRegistry.reg[t] = createResolver
	resolverRegistry.mu.Unlock()
}

func getResolverFor(t HostType) (func() Resolver, bool) {
	resolverRegistry.mu.Lock()
	r, ok := resolverRegistry.reg[t]
	resolverRegistry.mu.Unlock()
	return r, ok
}
