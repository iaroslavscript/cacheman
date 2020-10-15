package simplereplication

import (
	"log"
	"sync"
	"time"

	"github.com/iaroslavscript/cacheman/lib/common"
	"github.com/iaroslavscript/cacheman/lib/config"
)

type SimpleReplication struct {
	done  chan bool
	m     sync.RWMutex
	timer *time.Ticker

	CurrLog common.ReplLog
	NextLog common.ReplLog
	OldLog  common.ReplLog
}

func NewSimpleReplication(cfg *config.Config) *SimpleReplication {

	d := time.Duration(cfg.ReplicationRotateEveryMs) * time.Millisecond
	repl := SimpleReplication{
		done:  make(chan bool),
		timer: time.NewTicker(d),
	}

	repl.CurrLog.Info.Id = 2
	repl.CurrLog.Info.Id = 1
	repl.CurrLog.Info.Id = 0

	return &repl
}

func (s *SimpleReplication) Add(item common.ReplItem) {
	s.m.Lock()
	defer s.m.Unlock()

	s.NextLog.Data = append(s.CurrLog.Data, item)
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

	s.OldLog.Info.Id = s.CurrLog.Info.Id
	s.CurrLog.Info.Id = s.NextLog.Info.Id
	s.NextLog.Info.Id = s.NextLog.Info.Id + 1

	s.OldLog.Data = append(s.OldLog.Data, s.CurrLog.Data...)
	s.CurrLog.Data = make([]common.ReplItem, len(s.NextLog.Data))
	copy(s.CurrLog.Data, s.NextLog.Data)
	s.NextLog.Data = make([]common.ReplItem, 0)

	log.Printf("replication rotaion log %d ready. Size %d %d",
		s.CurrLog.Info.Id,
		len(s.OldLog.Data),
		len(s.CurrLog.Data),
	)

}
