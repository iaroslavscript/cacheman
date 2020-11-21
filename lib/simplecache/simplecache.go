package simplecache

import (
	"sync"

	"github.com/iaroslavscript/cacheman/lib/sdk"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const metricsSubsystem = "cache"

type SimpleCache struct {
	data                map[string]sdk.Record
	done                chan bool
	m                   sync.RWMutex
	opsApiRequestsTotal prometheus.Counter
	opsKeysTotal        prometheus.Gauge
	opsUsageBytes       prometheus.Gauge
}

func NewSimpleCache() *SimpleCache {
	c := SimpleCache{
		data: make(map[string]sdk.Record),
		done: make(chan bool),

		opsApiRequestsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: sdk.MetricsNamespace,
				Subsystem: metricsSubsystem,
				Name:      "api_requests_total",
				Help:      "The total number of requests to cache API",
			}),

		opsKeysTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: sdk.MetricsNamespace,
				Subsystem: metricsSubsystem,
				Name:      "keys_total",
				Help:      "The total number of keys stored in cache",
			}),

		opsUsageBytes: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: sdk.MetricsNamespace,
				Subsystem: metricsSubsystem,
				Name:      "cache_usage_bytes",
				Help:      "The size of cache in bytes",
			}),
	}

	c.opsApiRequestsTotal.Add(0.0)
	c.opsKeysTotal.Add(0.0)
	c.opsUsageBytes.Add(0.0)

	return &c
}

// Insert a new record or overwrite existed one.
func (c *SimpleCache) Insert(key sdk.KeyInfo, rec sdk.Record) {

	c.opsApiRequestsTotal.Inc()

	c.m.Lock()

	if _, ok := c.data[key.Key]; !ok { // Could we make it faster ???
		c.opsKeysTotal.Inc()
		c.opsUsageBytes.Add(0.0) // Curently we are not counting bytes
	}
	c.data[key.Key] = rec
	c.m.Unlock()
}

// Search for record equal to KeyInfo.Key which is not expired at the moment
// of KeyInfo.Expires
func (c *SimpleCache) Lookup(key sdk.KeyInfo) (sdk.Record, bool) {

	c.opsApiRequestsTotal.Inc()

	c.m.RLock()
	rec, ok := c.data[key.Key]
	c.m.RUnlock()

	if rec.Expires <= key.Expires {
		// the record will be expired at the requested moment key.Expires.
		// If key.Expires == time.Now().Unix() it means that the record
		// has already expired but the scheduler hasn't fired yet.
		ok = false
	}

	return rec, ok
}

// Delete record specified by key.Key
// If the record has been overwriten it will not be deleted
func (c *SimpleCache) Delete(key sdk.KeyInfo) {

	c.opsApiRequestsTotal.Inc()
	c.m.Lock()
	defer c.m.Unlock()

	if rec, ok := c.data[key.Key]; ok && rec.Expires <= key.Expires {
		// we need this check because record could have been overwriten
		// by new one and we don't need to delete it in that case.

		c.opsKeysTotal.Dec()
		c.opsUsageBytes.Set(0.0) // Curently we are not counting bytes
		delete(c.data, key.Key)
	}
}

// Reading records from chan and call Expired func.
// Should be run in a separete goroutine
func (c *SimpleCache) WatchSheduler(sched sdk.Scheduler) {
	for {
		select {
		case keyinfo := <-*sched.GetChan():
			c.Delete(keyinfo)
		case <-c.done:
			return
		}
	}
}

func (c *SimpleCache) Close() {
	c.done <- true
}
