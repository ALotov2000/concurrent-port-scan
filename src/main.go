package main

func main() {
	myServer, _ := NewServer() // todo: error handling
	forever := make(chan struct{})
	myServer.Launch()
	<-forever
}
