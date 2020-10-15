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

	"github.com/iaroslavscript/cacheman/lib/common"
	"github.com/iaroslavscript/cacheman/lib/config"
)

type Server struct {
	cache *common.Cache
	cfg   *config.Config
	repl  *common.Replication
	sched *common.Scheduler
}

func (s *Server) dataHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("request key:%s method:%s", r.URL.Path, r.Method)

	if r.URL.Path == "/" {

		switch r.Method {
		case http.MethodGet:
			s.healthHandler(w, r)
		case http.MethodHead:
			s.healthHandler(w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {

		switch r.Method {

		case http.MethodGet:
			s.lookupHandler(w, r)
		case http.MethodPost:
			s.insertHandler(w, r)
		case http.MethodHead:
			s.existsHandler(w, r)
		case http.MethodDelete:
			s.deleteHandler(w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("health request")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) existsHandler(w http.ResponseWriter, r *http.Request) {

	key := common.KeyInfo{
		Expires: time.Now().Unix(),
		Key:     r.URL.Path,
	}

	if _, ok := (*s.cache).Lookup(key); !ok {

		log.Printf("cache MISS %s", key.Key)
		http.NotFound(w, r)
		return
	}

	log.Printf("cache HIT %s", key.Key)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {

	key := common.KeyInfo{
		Expires: math.MaxInt64, // remove record regardless of it's expires date
		Key:     r.URL.Path,
	}

	log.Printf("DELETE key %s", key.Key)

	(*s.cache).Delete(key)

	w.WriteHeader(http.StatusOK)
}

func (s *Server) lookupHandler(w http.ResponseWriter, r *http.Request) {

	key := common.KeyInfo{
		Expires: time.Now().Unix(),
		Key:     r.URL.Path,
	}

	rec, ok := (*s.cache).Lookup(key)
	if !ok {

		log.Printf("cache MISS %s", key.Key)
		http.NotFound(w, r)
		return
	}

	log.Printf("cache HIT %s", key.Key)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(rec.Value)
}

func (s *Server) insertHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	var value []byte
	var err error
	var expires_in_sec int64

	if expires_in_sec, err = s.parseHeaderContentExpires(r); err != nil {

		log.Printf("error %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	expires := time.Now().Unix() + expires_in_sec

	if value, err = ioutil.ReadAll(r.Body); err != nil {
		log.Printf("Received incomplete %s, size %d", err, len(value))
	}

	log.Printf("insert %s expires_sec: %d", key, expires_in_sec)

	keyinfo := common.KeyInfo{
		Expires: expires,
		Key:     key,
	}

	rec := common.Record{
		Expires: expires,
		Value:   value,
	}

	(*s.cache).Insert(keyinfo, rec)
	(*s.repl).Add(*common.NewReplItem(0, keyinfo, rec))
	(*s.sched).Add(keyinfo)

	w.WriteHeader(http.StatusOK)
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

func NewServer(cfg *config.Config, cache common.Cache,
	repl common.Replication, sched common.Scheduler) *Server {

	return &Server{
		cache: &cache,
		cfg:   cfg,
		repl:  &repl,
		sched: &sched,
	}
}

func (s *Server) Serve() error {

	http.HandleFunc("/", s.dataHandler)

	log.Printf("server start listenning at %s", s.cfg.BindAddr)
	return http.ListenAndServe(s.cfg.BindAddr, nil)
}
