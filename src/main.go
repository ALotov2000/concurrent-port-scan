package main

type Server interface {
	Launch()
}

type Task interface {
	ID() int
	OpenPorts() chan uint16
	Handle()
}

func main() {
	server, _ := NewServer() // todo: error handling
	forever := make(chan struct{})
	server.Launch()
	<-forever
}
