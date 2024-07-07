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

type myTask struct {
	id        int
	ipAddr    string
	startPort int
	endPort   int
	openPorts chan uint16
}

func NewTask(ipAddr string, startPort int, endPort int) Task {

	lastID += 1
	log.Printf("myTask with id %d is created: Address=%s, ports=[%d, %d)", lastID, ipAddr, startPort, endPort)

	return &myTask{
		id:        lastID,
		ipAddr:    ipAddr,
		startPort: startPort,
		endPort:   endPort, // exclusive
		openPorts: make(chan uint16),
	}
}

func (t *myTask) Handle() {
	log.Printf("Handling task %d. Address: %s, Port Range: [%d, %d)", t.id, t.ipAddr, t.startPort, t.endPort)
	wg := &sync.WaitGroup{}
	for port := t.startPort; port < t.endPort; port += 1 {
		wg.Add(1)
		port := uint16(port)
		go func() {
			defer wg.Done()
			address := fmt.Sprintf("%s:%d", t.ipAddr, port)
			conn, err := net.DialTimeout("tcp", address, Timeout)
			if err == nil {
				_ = conn.Close()
				t.openPorts <- port
			}
		}()
	}
	wg.Wait()
	close(t.openPorts)
}

func (t *myTask) ID() int {
	return t.id
}

func (t *myTask) OpenPorts() chan uint16 {
	return t.openPorts
}
