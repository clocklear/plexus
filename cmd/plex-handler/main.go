package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	ph "github.com/clocklear/plex-handler/cmd/plex-handler/http"

	"github.com/go-kit/kit/log"
)

func main() {

	// Config.
	var (
		httpAddr  = flag.String("http.addr", ":3000", "HTTP listen address")
		debugAddr = flag.String("debug.addr", ":3001", "Debug and metrics listen address")
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

		// Server config
		{
			srv.Addr = *httpAddr
			srv.Handler = ph.DefaultRequestHandler(logger)
			srv.ReadTimeout = time.Second * 30
			srv.WriteTimeout = time.Second * 30
		}

		errc <- srv.ListenAndServe()
	}()

	// Run.
	logger.Log("exit", <-errc)
}
