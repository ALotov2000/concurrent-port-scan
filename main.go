package main

func main() {
	myServer := NewServer(5555, 10, 4096)
	forever := make(chan struct{})
	myServer.Launch()
	<-forever
}
