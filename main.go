package main

import (
	"flag"
	"log"
	"os"

	"github.com/iaroslavscript/cacheman/lib/config"
	"github.com/iaroslavscript/cacheman/lib/server"
	"github.com/iaroslavscript/cacheman/lib/simplecache"
	"github.com/iaroslavscript/cacheman/lib/simplereplication"
	"github.com/iaroslavscript/cacheman/lib/simplescheduler"
)

func parseFlag() {
	cfg := config.GetConfig()

	flag.StringVar(&cfg.BindAddr, "bind", cfg.BindAddr,
		"http server bind address.",
	)

	flag.Parse()
}

func main() {

	// get default config
	cfg := config.GetConfig()

	if _, err := os.Stat("/etc/cacheserver/config.json"); err == nil {
		// load config if exists
		if err = config.LoadConfig("/etc/cacheserver/config.json"); err != nil {
			log.Fatal("error loading config.json ", err.Error())
			os.Exit(1)
		}
	}

	// parse cmd arguments
	parseFlag()

	if err := config.Validate(); err != nil {
		log.Fatal("Config validation error ", err.Error())
		os.Exit(1)
	}

	cache := simplecache.NewSimpleCache()
	sched := simplescheduler.NewSimpleExpirer(cfg)
	repl := simplereplication.NewSimpleReplication(cfg)

	go repl.Start()
	go sched.Start()
	go cache.WatchSheduler(sched)

	defer func() {
		repl.Close()
		cache.Close()
		sched.Close()
	}()

	serv := server.NewServer(cfg, cache, repl, sched)

	log.Fatal(serv.Serve())
}
