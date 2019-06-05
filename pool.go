package connmgr

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	mu                    *sync.RWMutex
	hosts                 []*Host
	lastRTTStats          atomic.Value
	lastRTTStatsUpdatedAt int64
	n                     int64
}

func NewPool(hosts ...*Host) *Pool {
	if hosts == nil {
		hosts = make([]*Host, 0)
	}
	p := &Pool{
		mu:    &sync.RWMutex{},
		hosts: hosts,
	}

	p.sortedHosts(true)
	go p.maintain()
	return p
}

func (p *Pool) maintain() {
	for {
		p.sortedHosts(true)
		time.Sleep(1 * time.Second)
	}
}

func (p *Pool) Add(h *Host) {
	p.mu.Lock()
	p.hosts = append(p.hosts, h)
	p.mu.Unlock()
}

func (p *Pool) GetHost() *Host {
	return p.getLastFastestHost()
}

func (p *Pool) GetHostUncached() *Host {
	return p.getFastestHost(false)
}

func (p *Pool) GetNextHost() *Host {
	x := atomic.AddInt64(&p.n, 1)
	hosts := p.lastSortedHosts()
	if len(hosts) == 0 {
		return nil
	}
	return hosts[x%int64(len(hosts))].host
}

type rttStats struct {
	host *Host
	rtt  time.Duration
}

func (p *Pool) getFastestHost(forceUpdate bool) *Host {
	hosts := p.sortedHosts(forceUpdate)
	if len(hosts) == 0 {
		return nil
	}

	return hosts[0].host
}

func (p *Pool) getLastFastestHost() *Host {
	if nanotime()-atomic.LoadInt64(&p.lastRTTStatsUpdatedAt) > int64(10000*time.Millisecond) {
		return p.getFastestHost(true)
	}
	hosts, ok := p.lastRTTStats.Load().([]rttStats)
	if !ok || len(hosts) == 0 {
		return p.getFastestHost(true)
	}

	return hosts[0].host
}

func (p *Pool) lastSortedHosts() []rttStats {
	if nanotime()-atomic.LoadInt64(&p.lastRTTStatsUpdatedAt) > int64(10000*time.Millisecond) {
		return p.sortedHosts(true)
	}
	hosts, ok := p.lastRTTStats.Load().([]rttStats)
	if !ok || len(hosts) == 0 {
		return p.sortedHosts(true)
	}

	return hosts
}
func (p *Pool) sortedHosts(forceUpdate bool) []rttStats {
	p.mu.RLock()
	res := []rttStats{}
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for i, host := range p.hosts {
		wg.Add(1)
		go func(i int, host *Host) {
			rtt, err := host.RTT()
			mu.Lock()
			if err == nil {
				res = append(res, rttStats{
					host: host,
					rtt:  rtt,
				})
			}
			mu.Unlock()
			wg.Done()
		}(i, host)
	}
	wg.Wait()
	p.mu.RUnlock()
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].rtt < res[j].rtt
	})

	if forceUpdate || nanotime()-atomic.LoadInt64(&p.lastRTTStatsUpdatedAt) > int64(10000*time.Millisecond) {
		copiedStats := make([]rttStats, len(res))
		copy(copiedStats, res)
		p.lastRTTStats.Store(copiedStats)
		atomic.StoreInt64(&p.lastRTTStatsUpdatedAt, nanotime())
	}

	return res
}
