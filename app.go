package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewServer(port uint16) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: BuildRouter(),
	}
}

func BuildRouter() http.Handler {
	router := gin.Default()
	router.GET("/", handleHealthCheck)
	return router
}

func handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "Healthy Server!!")
}
