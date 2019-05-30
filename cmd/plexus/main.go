package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	ph "github.com/clocklear/plexus/cmd/plexus/http"
	"github.com/clocklear/plexus/pkg/plex"

	"github.com/go-kit/kit/log"
)

func main() {

	// Config.
	var (
		httpAddr       = flag.String("http.addr", ":3000", "HTTP listen address")
		debugAddr      = flag.String("debug.addr", ":3001", "Debug and metrics listen address")
		storeDirectory = flag.String("db.path", "./store", "The folder to be used as a JSON database for plexus")
		configFile     = flag.String("config.file", "config.json", "The trigger configuration file")
	)
	flag.Parse()

	// Logging.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// Interrupt.
	errc := make(chan error, 1)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Debug.
	go func() {
		logger := log.With(logger, "transport", "debug")
		logger.Log("addr", *debugAddr)
		errc <- http.ListenAndServe(*debugAddr, nil)
	}()

	// App.
	go func() {
		var srv http.Server

		logger := log.With(logger, "transport", "http")
		logger.Log("addr", *httpAddr)

		// Set up store
		s, err := plex.NewStore(*storeDirectory)
		if err != nil {
			errc <- err
		}

		// Load config
		cf, err := os.Open(*configFile)
		if err != nil {
			errc <- err
		}
		cfg, err := plex.NewConfig(cf)
		if err != nil {
			errc <- err
		}

		logger.Log("store", *storeDirectory, "config", *configFile)

		// Server config
		h, err := ph.DefaultRequestHandler(logger, s, cfg)
		if err != nil {
			errc <- err
		}
		srv.Addr = *httpAddr
		srv.Handler = h
		srv.ReadTimeout = time.Second * 30
		srv.WriteTimeout = time.Second * 30

		errc <- srv.ListenAndServe()
	}()

	// Run.
	logger.Log("exit", <-errc)
}
