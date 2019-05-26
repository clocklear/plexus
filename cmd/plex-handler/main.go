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

	ph "github.com/clocklear/plex-handler/cmd/plex-handler/http"
	"github.com/clocklear/plex-handler/pkg/plex"

	"github.com/go-kit/kit/log"
)

func main() {

	// Config.
	var (
		httpAddr      = flag.String("http.addr", ":3000", "HTTP listen address")
		debugAddr     = flag.String("debug.addr", ":3001", "Debug and metrics listen address")
		storeFile     = flag.String("store.file", "activity.json", "The file used to capture webhook activity.")
		maxStoreItems = flag.Int("store.maxitems", 100, "Maximum number of items in the activity store")
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

		f, err := os.OpenFile(*storeFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0655)
		defer f.Close()
		if err != nil {
			errc <- err
		}
		s, err := plex.NewActivityStore(f, *maxStoreItems)
		if err != nil {
			errc <- err
		}
		logger.Log("store", *storeFile)

		// Server config
		h, err := ph.DefaultRequestHandler(logger, s)
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
