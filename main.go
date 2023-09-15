package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	var addr string
	if s, exist := os.LookupEnv("ADDRESS"); exist {
		addr = s
	} else {
		addr = ":3567"
	}

	var period int

	if s, exist := os.LookupEnv("PERIOD"); exist {
		var err error

		period, err = strconv.Atoi(s)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		period = 10
	}

	storage := NewStorage()
	s := NewServer(addr, storage)

	log.Println("Running HTTP server in goroutine")
	go s.Run(ctx)

	log.Println("Running storage cleaning process in goroutine")
	done := make(chan struct{})
	go storage.CleaningProcess(ctx, time.Duration(period)*time.Second, done)

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
