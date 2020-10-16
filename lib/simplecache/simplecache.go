package simplecache

import (
	"sync"

	"github.com/iaroslavscript/cacheman/lib/common"
)

type SimpleCache struct {
	data map[string]common.Record
	done chan bool
	m    sync.RWMutex
}

func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		data: make(map[string]common.Record),
		done: make(chan bool),
	}
}

// Insert a new record or overwrite existed one.
func (c *SimpleCache) Insert(key common.KeyInfo, rec common.Record) {
	c.m.Lock()
	c.data[key.Key] = rec
	c.m.Unlock()
}

// Search for record equal to KeyInfo.Key which is not expired at the moment
// of KeyInfo.Expires
func (c *SimpleCache) Lookup(key common.KeyInfo) (common.Record, bool) {
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
func (c *SimpleCache) Delete(key common.KeyInfo) {

	c.m.Lock()

	if rec, ok := c.data[key.Key]; ok && rec.Expires <= key.Expires {
		// we need this check because record could have been overwriten
		// by new one and we don't need to delete it in that case.

		delete(c.data, key.Key)
	}

	c.m.Unlock()
}

// Reading records from chan and call Expired func.
// Should be run in a separete goroutine
func (c *SimpleCache) WatchSheduler(sched common.Scheduler) {
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
