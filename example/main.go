package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/agschwender/autoreload"
)

func main() {
	log.Printf("Starting application")

	// In a real production environment, you will likely want the
	// autoreload flag or environment variables to be defaulted to false
	// so that it is only enabled in the local environment.
	var shouldAutoReload bool
	flag.BoolVar(&shouldAutoReload, "autoreload", true, "enable autoreload")
	flag.Parse()

	server := &http.Server{Addr: ":8000"}

	if shouldAutoReload {
		log.Printf("Auto-reload is enabled")
		autoreload.New(
			autoreload.WithMaxAttempts(6),
			autoreload.WithOnReload(func() {
				log.Printf("Received change event, shutting down")
				server.Shutdown(context.Background())
			}),
		).Start()
	}

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
