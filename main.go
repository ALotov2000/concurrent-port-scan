package main

import (
	"log"
	"net/http"
)

func main() {
	forever := make(chan struct{})
	server := NewServer(5555)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Application could not be created")
		}
	}()
	log.Printf("server has launched on: %s\n", server.Addr)

	<-forever
}
