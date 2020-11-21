package simplereplication

import (
	"log"
	"sync"
	"time"

	"github.com/iaroslavscript/cacheman/lib/config"
	"github.com/iaroslavscript/cacheman/lib/sdk"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const metricsSubsystem = "repl"
const binlogTypeOld = "old"

type SimpleReplication struct {
	done  chan bool
	m     sync.RWMutex
	timer *time.Ticker

	CurrLog sdk.ReplLog
	NextLog sdk.ReplLog
	OldLog  sdk.ReplLog

	opsApiRequestsTotal   prometheus.Counter
	opsBinLogsTotal       prometheus.Counter
	opsBinLogRecordsTotal *prometheus.CounterVec
	opsBinLogBytes        *prometheus.CounterVec
}

func NewSimpleReplication(cfg *config.Config) *SimpleReplication {

	d := time.Duration(cfg.ReplicationRotateEveryMs) * time.Millisecond
	repl := SimpleReplication{
		done: make(chan bool),

		opsApiRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: sdk.MetricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "api_requests_total",
			Help:      "The total number of requests to replication API",
		}),

		opsBinLogsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: sdk.MetricsNamespace,
				Subsystem: metricsSubsystem,
				Name:      "binlogs_total",
				Help:      "The total number of binary logs",
			}),

		opsBinLogRecordsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: sdk.MetricsNamespace,
				Subsystem: metricsSubsystem,
				Name:      "binlog_records_total",
				Help:      "The total number of records in binary logs grouped by log type",
			},
			[]string{"type"},
		),

		opsBinLogBytes: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: sdk.MetricsNamespace,
				Subsystem: metricsSubsystem,
				Name:      "binlog_bytes",
				Help:      "The size of binary logs in bytes grouped by log type",
			},
			[]string{"type"},
		),

		timer: time.NewTicker(d),
	}

	repl.CurrLog.Info.Id = 2
	repl.CurrLog.Info.Id = 1
	repl.CurrLog.Info.Id = 0

	// Init counter
	repl.opsApiRequestsTotal.Add(0.0)
	repl.opsBinLogsTotal.Add(0.0)
	repl.opsBinLogRecordsTotal.WithLabelValues(binlogTypeOld).Add(0.0)
	repl.opsBinLogBytes.WithLabelValues(binlogTypeOld).Add(0.0)

	return &repl
}

// TODO use chan here to avoid blocking on adding
func (s *SimpleReplication) Add(item sdk.ReplItem) {

	s.opsApiRequestsTotal.Inc()

	s.m.Lock()
	defer s.m.Unlock()

	s.NextLog.Data = append(s.NextLog.Data, item)
}

func (s *SimpleReplication) Start() {

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

func (s *SimpleReplication) Close() {
	s.done <- true
}

func (s *SimpleReplication) tick() {

	s.m.Lock()
	defer s.m.Unlock()

	nextlog_n := len(s.NextLog.Data)

	if nextlog_n == 0 {
		return
	}

	s.OldLog.Info.Id++
	s.CurrLog.Info.Id++
	s.NextLog.Info.Id++

	currlog_n := len(s.CurrLog.Data)

	s.OldLog.Data = append(s.OldLog.Data, s.CurrLog.Data...)
	s.CurrLog.Data = make([]sdk.ReplItem, len(s.NextLog.Data))
	copy(s.CurrLog.Data, s.NextLog.Data)

	// the next log could be at least as big as it was before
	s.NextLog.Data = make([]sdk.ReplItem, 0, nextlog_n)

	s.opsBinLogRecordsTotal.WithLabelValues(binlogTypeOld).Add(float64(currlog_n))

	// Curently we are not counting bytes
	s.opsBinLogBytes.WithLabelValues(binlogTypeOld).Add(0.0)

	// it's better to unlock here or make a copy of nextlog and unlock
	// because we add only to next not to curr and old
	// but in that case we need two mutex (next_mutex and curr_old_mutex)

	s.opsBinLogsTotal.Inc()
	log.Printf("replication_log:%d buckets_sizes:[%d, %d]",
		s.CurrLog.Info.Id,
		len(s.OldLog.Data),
		len(s.CurrLog.Data),
	)

}
