package main

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	storage := NewStorage()
	s := NewServer(":3000", storage)
	log.Println("Running HTTP server in goroutine")
	go s.Run(ctx)
	done := make(chan struct{})

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
