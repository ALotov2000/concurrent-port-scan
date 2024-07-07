package main

func main() {
	myServer := NewServer(5555, 101, 32)
	forever := make(chan struct{})
	myServer.Launch()
	<-forever
}
