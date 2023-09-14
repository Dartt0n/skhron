package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	storage := NewStorage()
	s := NewServer(":3000", storage)

	log.Println("Running HTTP server in goroutine")
	go s.Run(ctx)

	log.Println("Running storage cleaning process in goroutine")
	done := make(chan struct{})
	go storage.CleaningProcess(ctx, done)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)

	log.Println("Waiting for Ctrl-C to terminate...")
	<-c
	log.Println("Closing all processes...")
	cancel()
	<-done
	s.Shutdown(ctx)
}
