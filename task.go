package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

var lastID = 0

const MinPortNum = 0     // inclusive
const MaxPortNum = 65536 // exclusive
const Timeout = time.Millisecond * time.Duration(1000)

type Task struct {
	ID        int
	ipAddr    string
	startPort int
	endPort   int
	OpenPorts chan int
}

func NewTask(ipAddr string, startPort int, endPort int) *Task {

	lastID += 1
	log.Printf("Task with id %d is created: Address=%s, ports=[%d, %d)", lastID, ipAddr, startPort, endPort)

	return &Task{
		ID:        lastID,
		ipAddr:    ipAddr,
		startPort: startPort,
		endPort:   endPort, // exclusive
		OpenPorts: make(chan int),
	}
}

func (t *Task) Handle() {
	log.Printf("Handling task %d. Address: %s, Port Range: [%d, %d)", t.ID, t.ipAddr, t.startPort, t.endPort)
	wg := &sync.WaitGroup{}
	for port := t.startPort; port < t.endPort; port += 1 {
		wg.Add(1)
		port := port
		go func() {
			defer wg.Done()
			address := fmt.Sprintf("%s:%d", t.ipAddr, port)
			conn, err := net.DialTimeout("tcp", address, Timeout)
			if err == nil {
				_ = conn.Close()
				t.OpenPorts <- port
			}
		}()
	}
	wg.Wait()
	close(t.OpenPorts)
}
