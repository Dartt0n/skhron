# Skhron
![go workflow](https://github.com/dartt0n/skhron/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/dartt0n/skhron)](https://goreportcard.com/report/github.com/dartt0n/skhron)

Skhron is a lightweight in-memory storage designed for in-code usage

# Usage

## Simple Storage Example

> [Example Source Code](./examples/simple_storage)

```go
package main

import (
	"fmt"
	"time"

	"github.com/dartt0n/skhron"
)

func main() {
	storage := skhron.New() // init new storage
	storage.LoadSnapshot()  // load snapshot from file (default: `./.skhron/snapshot.skh`)

	storage.CleanUp() // manually run cleaup - remove expored keys

	timestamp := time.Now().Format("2006_01_02 15:04:05")

	// .Exists(key) checks whether the key exists in the storage
	if storage.Exists("run-timestamp") {
		fmt.Printf("Previous run timestamp found!\n")
	}

	// .Get(key) returns value associated with the key
	if value, err := storage.Get("run-timestamp"); err != nil {
		fmt.Printf("Get failed: %v\n", err)
	} else {
		fmt.Printf("Value: %s\n", string(value)) // convert bytes to string and print
	}

	// .Delete(key) deletes the key from the storage.
	// This method call will NOT panic or return error if key does not exist
	if err := storage.Delete("run-timestamp"); err != nil {
		fmt.Printf("Delete failed: %v\n", err)
	}

	// .Put(key, value, ttl) puts value under key with time to live equal to ttl
	// Here we put bytes of string `timestamp` under the key `run-timestamp` and time-to-live equal to 1 hour
	if err := storage.PutTTL("run-timestamp", []byte(timestamp), 1*time.Hour); err != nil {
		fmt.Printf("Put failed: %v\n", err)
	}
}
```

## Example with HTTP server

> [Example Source Code](./examples/http_server)

```go
func main() {
	ctx, cancel := context.WithCancel(context.Background())

	addr := flag.String("address", ":3567", "the address to listen on")
	period := flag.Int("period", 10, "the period of time to run cleanup (in seconds)")

	flag.Parse()

	storage := skhron.New() // initialize storage
	storage.LoadSnapshot() // load from snapshot

	server := newServer(*addr, storage) // basic http server

	go server.Run(ctx) // run server in background

	storage_shutdown := make(chan struct{}) // create channel for shutdown for storage
    
    // run periodic cleanup process in background.
    // it would perform cleanup each `period` seconds.
    // it would finishon ctx.cancel() call and put signal into `storage_shutdown` channel 
	go storage.PeriodicCleanup(ctx, time.Duration(*period)*time.Second, storage_shutdown)

    // listen for termination signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)

    <-c // wait for Ctrl-C
	cancel() // send cancelation signal to both http server and storage process

	server.Shutdown(ctx) // wait for server

	<-storage_shutdown // wait for storage
}

```