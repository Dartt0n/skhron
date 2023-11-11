package main

import (
	"fmt"
	"time"

	"github.com/dartt0n/skhron"
)

func main() {
	storage := skhron.New[string]() // init new storage
	storage.LoadSnapshot()          // load snapshot from file (default: `./.skhron/snapshot.skh`)

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
		fmt.Printf("Value: %s\n", value) // convert bytes to string and print
	}

	// .Delete(key) deletes the key from the storage.
	// This method call will NOT panic or return error if key does not exist
	if err := storage.Delete("run-timestamp"); err != nil {
		fmt.Printf("Delete failed: %v\n", err)
	}

	// .Put(key, value, ttl) puts value under key with time to live equal to ttl
	// Here we put bytes of string `timestamp` under the key `run-timestamp` and time-to-live equal to 1 hour
	if err := storage.Put("run-timestamp", timestamp, 1*time.Hour); err != nil {
		fmt.Printf("Put failed: %v\n", err)
	}
}
