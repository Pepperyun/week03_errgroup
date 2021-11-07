package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)

	g, groupCtx := errgroup.WithContext(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	g.Go(func() error {
		fmt.Println("starting server")
		return server.ListenAndServe()
	})

	g.Go(func() error {
		<-groupCtx.Done()

		fmt.Println("shutting down server...")
		return server.Shutdown(groupCtx)
	})

	g.Go(func() error {
		chanel := make(chan os.Signal, 1)
		signal.Notify(chanel, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-groupCtx.Done():
			return groupCtx.Err()
		case <-chanel:
			cancel()
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Println("group error: ", err)
	}
	fmt.Println("all group exit")
}
