package main

type Server interface {
	Launch()
}

type Task interface {
	ID() int
	OutputChannel() chan Output
	Handle()
}

type Output interface {
}

func main() {
	server, _ := NewServer() // todo: error handling
	forever := make(chan struct{})
	server.Launch()
	<-forever
}
