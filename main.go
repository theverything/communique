package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theverything/communique/internal/dispatch"
	"github.com/theverything/communique/internal/server"
	"github.com/theverything/communique/internal/store"
)

func main() {
	shutdown := make(chan struct{}, 1)
	ctx := context.Background()
	ctxc, cancel := context.WithCancel(ctx)
	s := store.New()
	d := dispatch.New(ctxc, dispatch.DispatcherConfig{Concurrency: 5}, s)

	d.Start()

	log.Println("server starting on port 8080")
	srv := server.New(server.Config{Port: 8080}, d)

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
	d.Wait()

	srv.Shutdown(ctxt)
}
