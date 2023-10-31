package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dartt0n/skhron/server"
	"github.com/dartt0n/skhron/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	addr := flag.String("address", ":3567", "the address to listen on")
	period := flag.Int("period", 10, "the period of time to run cleanup (in seconds)")

	flag.Parse()

	storage := storage.New()
	server := server.New(*addr, storage)

	log.Println("Running HTTP server in goroutine")
	go server.Run(ctx)

	log.Println("Running storage cleaning process in goroutine")
	done := make(chan struct{})
	go storage.CleaningProcess(ctx, time.Duration(*period)*time.Second, done)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)

	log.Println("Waiting for Ctrl-C to terminate...")
	<-c
	log.Println("Closing all processes...")
	cancel()
	<-done
	server.Shutdown(ctx)
}
