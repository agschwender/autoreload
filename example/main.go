package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/agschwender/autoreload"
)

func main() {
	log.Printf("Starting application")

	server := &http.Server{Addr: ":8000"}

	autoreload.New(
		autoreload.WithMaxAttempts(6),
		autoreload.WithOnReload(func() {
			server.Shutdown(context.Background())
		}),
	).Start()

	go func() {
		log.Printf("Starting HTTP server")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt, syscall.SIGTERM)
	<-interruptChannel

	log.Printf("Received interrupt signal, shutting down")
	server.Shutdown(context.Background())
}
