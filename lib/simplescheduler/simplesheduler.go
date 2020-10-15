package simplescheduler

import (
	"container/heap"
	"sync"
	"time"

	"github.com/iaroslavscript/cacheman/lib/common"
	"github.com/iaroslavscript/cacheman/lib/config"
)

type SimpleExpirer struct {
	C         chan common.KeyInfo
	cfg       *config.Config
	done      chan bool
	m         sync.Mutex
	timer     *time.Ticker
	timetable SchedMinHeap
}

func NewSimpleExpirer(cfg *config.Config) *SimpleExpirer {

	d := time.Duration(cfg.ShedulerDelExpiredEverySec) * time.Second
	x := &SimpleExpirer{
		C:         make(chan common.KeyInfo, cfg.ShedulerExpiredQuequeSize),
		cfg:       cfg,
		done:      make(chan bool),
		timer:     time.NewTicker(d),
		timetable: make(SchedMinHeap, 0),
	}

	heap.Init(&(x.timetable))

	return x
}

func (s *SimpleExpirer) Add(key common.KeyInfo) {

	// round expires time to the next sheduler tick
	expires := (key.Expires/s.cfg.ShedulerDelExpiredEverySec + 1) * s.cfg.ShedulerDelExpiredEverySec

	s.m.Lock()
	defer s.m.Unlock()

	s.timetable.Push(&schedHeapItem{
		value:    key.Key,
		priority: expires,
	})
}

func (s *SimpleExpirer) Start() {

	defer s.timer.Stop()

	for {
		select {
		case <-s.timer.C:
			s.tick()
		case <-s.done:
			return
		}
	}
}

func (s *SimpleExpirer) Close() {
	s.done <- true
}

func (s *SimpleExpirer) GetChan() *chan common.KeyInfo {
	return &s.C
}

func (s *SimpleExpirer) tick() {

	s.m.Lock()
	defer s.m.Unlock()

	t := time.Now().Unix()
	for (s.timetable.Len() > 0) && (s.timetable[0].priority <= t) {
		item := heap.Pop(&s.timetable).(*schedHeapItem)

		s.C <- common.KeyInfo{
			Expires: item.priority,
			Key:     item.value,
		}
	}
}
