package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
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
	router.GET("/:domain", myServer.handleQuery)
	myServer.HttpServer.Handler = router
}

func (myServer *MyServer) handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "Healthy Server!!")
}

func (myServer *MyServer) handleQuery(c *gin.Context) {
	if err := validateQuery(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("failure: %v", err),
		})
		return
	}

	domain := c.Param("domain")
	fromPort := getFromPort(c)
	toPort := getToPort(c)

	var tasks []Task
	for startPort := fromPort; startPort < toPort; startPort += myServer.portChunkSize {
		newTask := *NewTask(
			domain,
			startPort,
			int(math.Min(float64(startPort+myServer.portChunkSize), float64(toPort))),
		)
		tasks = append(tasks, newTask)
		myServer.TaskQueue <- newTask
	}

	var ports []int
	for _, task := range tasks {
		for openPort := range task.OpenPorts {
			ports = append(ports, openPort)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"domain":  domain,
		"ports":   ports,
	})
}

func validateQuery(c *gin.Context) []error {
	var errorList []error
	domain := c.Param("domain")
	fromPortStr := c.Query("from")
	toPortStr := c.Query("to")

	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]+\.)*[a-zA-Z0-9]+\.[a-zA-Z]{2,}$`)
	if !ipRegex.MatchString(domain) && !domainRegex.MatchString(domain) {
		errorList = append(errorList, errors.New("invalid domain"))
	}

	_, err := strconv.Atoi(fromPortStr)
	if fromPortStr != "" && err != nil {
		errorList = append(errorList, errors.New("bad 'from' parameter"))
	}

	_, err = strconv.Atoi(toPortStr)
	if fromPortStr != "" && err != nil {
		errorList = append(errorList, errors.New("bad 'to' parameter"))
	}

	return errorList
}

func getFromPort(c *gin.Context) int {
	fromPortStr := c.Query("from")
	if fromPortStr == "" {
		return MinPortNum
	}
	fromPort, _ := strconv.Atoi(fromPortStr)
	return fromPort
}

func getToPort(c *gin.Context) int {
	toPortStr := c.Query("to")
	if toPortStr == "" {
		return MaxPortNum
	}
	toPort, _ := strconv.Atoi(toPortStr)
	return toPort
}
