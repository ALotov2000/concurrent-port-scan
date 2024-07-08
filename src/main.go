package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server interface {
	Launch(ctx *context.Context)
}

type Task interface {
	ID() int
	OutputChannel() chan Output
	Handle()
	WaitGroup() *sync.WaitGroup
}

type Output interface {
}

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop,
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	ctx, cancel := context.WithCancel(context.Background())
	server, _ := NewServer() // todo: error handling
	server.Launch(&ctx)

	s := <-stop
	log.Printf("initiating graceful shutdown: (got signal: %v)\n", s)
	cancel()

	<-time.After(3 * time.Second)
}
