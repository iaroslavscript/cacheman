package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/iaroslavscript/cacheman/lib/config"
	"github.com/iaroslavscript/cacheman/lib/sdk"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const metricsSubsystem = "server"

type Server struct {
	cache               *sdk.Cache
	cfg                 *config.Config
	repl                *sdk.Replication
	sched               *sdk.Scheduler
	opsApiRequestsTotal prometheus.Counter
}

func pathToKey(path string) string {
	if len(path) > 1 {
		return path[1:]
	}

	return ""
}

func requestInfo(start time.Time, code int, r *http.Request,
	f string, args ...interface{}) string {

	elapsed := time.Now().Sub(start)

	return fmt.Sprintf("request:%s method:%s from:%s response:%d response_time:%d %s",
		r.URL.Path,
		r.Method,
		r.Host,
		code,
		elapsed.Milliseconds(),
		fmt.Sprintf(f, args...),
	)
}

func (s *Server) dataHandler(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	s.opsApiRequestsTotal.Inc()

	if r.URL.Path == "/" {

		switch r.Method {
		case http.MethodGet:
			s.healthHandler(start, w, r)
		case http.MethodHead:
			s.healthHandler(start, w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
			log.Printf(requestInfo(start, http.StatusBadRequest, r, ""))
		}
	} else {

		switch r.Method {

		case http.MethodGet:
			s.lookupHandler(start, w, r)
		case http.MethodPost:
			s.insertHandler(start, w, r)
		case http.MethodHead:
			s.existsHandler(start, w, r)
		case http.MethodDelete:
			s.deleteHandler(start, w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
			log.Printf(requestInfo(start, http.StatusBadRequest, r, ""))
		}
	}
}

func (s *Server) healthHandler(t time.Time, w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	log.Printf(requestInfo(t, http.StatusOK, r, ""))
}

func (s *Server) existsHandler(t time.Time, w http.ResponseWriter, r *http.Request) {

	key := sdk.KeyInfo{
		Expires: time.Now().Unix(),
		Key:     pathToKey(r.URL.Path),
	}

	if _, ok := (*s.cache).Lookup(key); !ok {

		http.NotFound(w, r)
		log.Printf(requestInfo(t, http.StatusNotFound, r, ""))
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf(requestInfo(t, http.StatusOK, r, ""))
}

func (s *Server) parseHeaderContentExpires(r *http.Request) (int64, error) {
	var expires_in_sec int64
	var e error

	if val := r.Header.Get("X-Content-Expires-At"); val != "" {
		// no-op
		// TODO for next version
	} else if val = r.Header.Get("X-Content-Expires-Sec"); val != "" {

		valint, err := strconv.Atoi(val)
		if err != nil {
			e = errors.New(fmt.Sprintf("Improper value of %s http header",
				"X-Content-Expires-Sec",
			))
			return expires_in_sec, e
		}

		expires_in_sec = int64(valint)
	} else {

		expires_in_sec = s.cfg.ExpiresDefaultDurationSec
	}

	if expires_in_sec < 1 {
		e = errors.New(fmt.Sprintf("Improper value of %s or %s HTTP header",
			"X-Content-Expires-Sec",
			"X-Content-Expires-At",
		))
	}

	return expires_in_sec, e
}

func (s *Server) deleteHandler(t time.Time, w http.ResponseWriter, r *http.Request) {

	key := sdk.KeyInfo{
		Expires: math.MaxInt64, // remove record regardless of it's expires date
		Key:     pathToKey(r.URL.Path),
	}

	// We need to delete only if keys was found to avoid overpopulation of
	// replication log's bucket.
	// It's better to move it to cache or because we missing sheduler deletion
	//(*s.repl).Delete(*sdk.NewReplItem(0, keyinfo, rec))

	(*s.cache).Delete(key)
	w.WriteHeader(http.StatusOK)
	log.Printf(requestInfo(t, http.StatusOK, r, ""))
}

func (s *Server) lookupHandler(t time.Time, w http.ResponseWriter, r *http.Request) {

	key := sdk.KeyInfo{
		Expires: time.Now().Unix(),
		Key:     pathToKey(r.URL.Path),
	}

	rec, ok := (*s.cache).Lookup(key)
	if !ok {

		http.NotFound(w, r)
		log.Printf(requestInfo(t, http.StatusNotFound, r, ""))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(rec.Value)
	log.Printf(requestInfo(t, http.StatusOK, r, "value_size:%d", len(rec.Value)))
}

func (s *Server) insertHandler(t time.Time, w http.ResponseWriter, r *http.Request) {
	key := pathToKey(r.URL.Path)
	var value []byte
	var err error
	var expires_in_sec int64

	if expires_in_sec, err = s.parseHeaderContentExpires(r); err != nil {

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		log.Printf(requestInfo(t, http.StatusBadRequest, r, "error:%s", err.Error()))
		return
	}

	expires := time.Now().Unix() + expires_in_sec

	if value, err = ioutil.ReadAll(r.Body); err != nil {
		value_n := len(value)
		msg := fmt.Sprintf("Received incomplete %s, size %d",
			err.Error(),
			value_n,
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		log.Printf(requestInfo(t, http.StatusBadRequest, r,
			"error:'%s' size: %d",
			msg,
			value_n,
		))
		return
	}

	keyinfo := sdk.KeyInfo{
		Expires: expires,
		Key:     key,
	}

	rec := sdk.Record{
		Expires: expires,
		Value:   value,
	}

	(*s.cache).Insert(keyinfo, rec)
	(*s.repl).Add(*sdk.NewReplItem(0, keyinfo, rec))
	(*s.sched).Add(keyinfo)

	log.Printf(requestInfo(t, http.StatusOK, r, "expires_sec:%d", expires_in_sec))
	w.WriteHeader(http.StatusOK)
}

func NewServer(cfg *config.Config, cache sdk.Cache,
	repl sdk.Replication, sched sdk.Scheduler) *Server {

	s := Server{
		cache: &cache,
		cfg:   cfg,
		repl:  &repl,
		sched: &sched,
		opsApiRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: sdk.MetricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "api_requests_total",
			Help:      "The total number of processed events",
		}),
	}

	s.opsApiRequestsTotal.Add(0.0)

	return &s
}

func (s *Server) Serve() error {

	http.HandleFunc("/", s.dataHandler)

	log.Printf("server start listenning at %s", s.cfg.BindAddr)
	return http.ListenAndServe(s.cfg.BindAddr, nil)
}
