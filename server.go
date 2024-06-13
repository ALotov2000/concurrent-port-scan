package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"net/http"
	"sync"
)

type MyServer struct {
	HttpServer    http.Server
	TaskQueue     chan Task
	numWorker     int
	numChunk      int
	portChunkSize int
}

func NewServer(port uint16, numWorker int, portChunkSize int) *MyServer {
	taskQueue := make(chan Task, 2*numWorker)
	myNewServer := &MyServer{
		HttpServer: http.Server{
			Addr: fmt.Sprintf(":%d", port),
		},
		TaskQueue:     taskQueue,
		numWorker:     numWorker,
		portChunkSize: portChunkSize,
	}

	myNewServer.setupRouter()

	return myNewServer
}

func (myServer *MyServer) Launch() {
	go myServer.launchHttp()
	go myServer.launchApplication()
}

func (myServer *MyServer) launchHttp() {
	log.Printf("launching server on: %s\n", myServer.HttpServer.Addr)
	if err := myServer.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Application could not be created")
	}
}

func (myServer *MyServer) launchApplication() {
	log.Printf("launching port scan application: %s\n", myServer.HttpServer.Addr)
	wg := &sync.WaitGroup{}
	for i := 1; i <= myServer.numWorker; i++ {
		go worker(i, myServer.TaskQueue, wg)
	}
	wg.Wait()
}

func (myServer *MyServer) setupRouter() {
	router := gin.Default()
	router.GET("/", myServer.handleHealthCheck)
	router.GET("/:ipAddr", myServer.handleQuery)
	myServer.HttpServer.Handler = router
}

func (myServer *MyServer) handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "Healthy Server!!")
}

func (myServer *MyServer) handleQuery(c *gin.Context) {
	ipAddress := c.Param("ipAddr")
	var tasks []Task
	for startPort := 0; startPort < 65536; startPort += myServer.portChunkSize {
		newTask := *NewTask(
			ipAddress,
			startPort,
			int(math.Min(float64(startPort+myServer.portChunkSize), 65536)),
		)
		tasks = append(tasks, newTask)
		myServer.TaskQueue <- newTask
	}
	var ports []int
	for _, task := range tasks {
		for result := range task.Results {
			ports = append(ports, result.Ports...)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"ipAddress": ipAddress,
		"ports":     ports,
	})
}
