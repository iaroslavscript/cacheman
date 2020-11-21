package simplescheduler

import (
	"container/heap"
	"sync"
	"time"

	"github.com/iaroslavscript/cacheman/lib/config"
	"github.com/iaroslavscript/cacheman/lib/sdk"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const metricsSubsystem = "sched"

type SimpleExpirer struct {
	C                   chan sdk.KeyInfo
	cfg                 *config.Config
	done                chan bool
	m                   sync.Mutex
	opsApiRequestsTotal prometheus.Counter
	opsRecsTotal        prometheus.Gauge
	opsTriggeredTotal   prometheus.Counter
	timer               *time.Ticker
	timetable           SchedMinHeap
}

func NewSimpleExpirer(cfg *config.Config) *SimpleExpirer {

	d := time.Duration(cfg.ShedulerDelExpiredEverySec) * time.Second
	x := &SimpleExpirer{
		C:    make(chan sdk.KeyInfo, cfg.ShedulerExpiredQuequeSize),
		cfg:  cfg,
		done: make(chan bool),

		opsApiRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: sdk.MetricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "api_requests_total",
			Help:      "The total number of requests to scheduler API",
		}),

		opsRecsTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: sdk.MetricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "records_total",
			Help:      "The number of records are sheduled for expiring",
		}),

		opsTriggeredTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: sdk.MetricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "triggered_total",
			Help:      "The number of times sheduled is triggered",
		}),
		timer:     time.NewTicker(d),
		timetable: make(SchedMinHeap, 0),
	}

	heap.Init(&(x.timetable))

	x.opsApiRequestsTotal.Add(0.0)
	x.opsRecsTotal.Set(0.0)
	x.opsTriggeredTotal.Add(0.0)

	return x
}

func (s *SimpleExpirer) Add(key sdk.KeyInfo) {

	s.opsApiRequestsTotal.Inc()
	s.opsRecsTotal.Inc()

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

func (s *SimpleExpirer) GetChan() *chan sdk.KeyInfo {
	return &s.C
}

func (s *SimpleExpirer) tick() {

	s.opsTriggeredTotal.Inc()

	s.m.Lock()
	defer s.m.Unlock()

	t := time.Now().Unix()
	for (s.timetable.Len() > 0) && (s.timetable[0].priority <= t) {
		item := heap.Pop(&s.timetable).(*schedHeapItem)

		s.C <- sdk.KeyInfo{
			Expires: item.priority,
			Key:     item.value,
		}

		s.opsRecsTotal.Dec()
	}
}
