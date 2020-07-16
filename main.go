package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theverything/communique/internal/hub"
	"github.com/theverything/communique/internal/server"
	"github.com/theverything/communique/internal/store"
)

var disableCORS = false

func main() {
	flag.BoolVar(&disableCORS, "disable-cors", false, "Disable CORS headers")

	flag.Parse()

	shutdown := make(chan struct{}, 1)
	ctx := context.Background()
	ctxc, cancel := context.WithCancel(ctx)
	s := store.New()
	d := hub.New(ctxc, hub.Config{Concurrency: 5}, s)

	go d.Start()

	if disableCORS {
		log.Println("CORS disabled")
	}

	log.Println("server starting on port 8080")
	srv := server.New(server.Config{Port: 8080, DisableCORS: disableCORS}, d)

	stop := make(chan os.Signal, 1)
	signal.Notify(
		stop,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)

	go func() {
		<-stop
		shutdown <- struct{}{}
	}()

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Println(err)
			shutdown <- struct{}{}
		}
	}()

	<-shutdown

	ctxt, cancelT := context.WithTimeout(ctx, time.Second*5)
	defer cancelT()

	cancel()

	srv.Shutdown(ctxt)
}
