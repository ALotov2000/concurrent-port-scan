package main

import (
	"context"
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

type myServer struct {
	httpServer http.Server
	config     *config
}

type config struct {
	port          uint
	portChunkSize uint
	clusters      []cluster
}

type cluster struct {
	mean      uint
	numWorker uint
	taskQueue chan Task
}

func NewServer() (Server, error) {
	configObj, _ := getConfigObj() // todo: error handling
	myNewServer := &myServer{
		httpServer: http.Server{
			Addr: fmt.Sprintf(":%d", configObj.port),
		},
		config: configObj,
	}

	myNewServer.setupRouter()

	return myNewServer, nil
}

func getConfigObj() (*config, error) { // todo: must be read from a yml file
	return &config{
		port:          5555,
		portChunkSize: 32,
		clusters: []cluster{
			{mean: 1000, numWorker: 70, taskQueue: make(chan Task)},
			{mean: 10000, numWorker: 30, taskQueue: make(chan Task)},
			{mean: 40000, numWorker: 30, taskQueue: make(chan Task)},
		},
	}, nil
}

func (myServer *myServer) Launch(ctx *context.Context) {
	go myServer.launchHttp()
	go myServer.launchApplication(ctx)
}

func (myServer *myServer) launchHttp() {
	log.Printf("launching server on: %s\n", myServer.httpServer.Addr)
	if err := myServer.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Application could not be created")
	}
}

func (myServer *myServer) launchApplication(ctx *context.Context) {
	log.Printf("launching port scan application: %s\n", myServer.httpServer.Addr)
	var wg sync.WaitGroup
	var id uint = 1
	for _, c := range myServer.config.clusters {
		for i := 1; i <= int(c.numWorker); i++ {
			wg.Add(1)
			go worker(id, c.taskQueue, &wg, ctx)
			id += 1
		}
	}
	<-(*ctx).Done()
	wg.Wait()
}

func (myServer *myServer) setupRouter() {
	router := gin.Default()
	router.GET("/", myServer.handleHealthCheck)
	router.GET("/:domain", myServer.handleQuery)
	myServer.httpServer.Handler = router
}

func (myServer *myServer) handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "Healthy Server!!")
}

func (myServer *myServer) handleQuery(c *gin.Context) {
	if err := validateQuery(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("failure: %v", err),
		})
		return
	}
	domain, fromPort, toPort := myServer.getParameters(c)
	outputChannel := make(chan Output)
	var wg sync.WaitGroup
	for startPort := fromPort; startPort <= toPort; startPort += myServer.config.portChunkSize {
		wg.Add(1)
		newTask := NewTask(
			domain,
			startPort,
			uint(math.Min(float64(startPort+myServer.config.portChunkSize), float64(toPort))),
			outputChannel,
			&wg,
		)
		c := myServer.findBestCluster(startPort)
		c.taskQueue <- newTask
	}
	go func() {
		wg.Wait()
		close(outputChannel)
	}()
	myServer.printOutputs(c, outputChannel)
}

func (myServer *myServer) getParameters(c *gin.Context) (string, uint, uint) {
	domain := c.Param("domain")
	fromPort := getFromPort(c)
	toPort := getToPort(c)
	return domain, fromPort, toPort
}

func (myServer *myServer) printOutputs(c *gin.Context, outputChannel chan Output) {
	result := make([]Output, 0)
	for o := range outputChannel {
		result = append(result, o)
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

func (myServer *myServer) findBestCluster(startPort uint) cluster {
	bestD := math.Inf(1)
	var bestC cluster
	for _, c := range myServer.config.clusters {
		d := math.Abs(float64(c.mean - startPort))
		if d < bestD {
			bestD = d
			bestC = c
		}
	}
	return bestC
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

func getFromPort(c *gin.Context) uint {
	fromPortStr := c.Query("from")
	if fromPortStr == "" {
		return MinPortNum
	}
	fromPort, _ := strconv.Atoi(fromPortStr)
	return uint(fromPort)
}

func getToPort(c *gin.Context) uint {
	toPortStr := c.Query("to")
	if toPortStr == "" {
		return MaxPortNum
	}
	toPort, _ := strconv.Atoi(toPortStr)
	return uint(toPort)
}
