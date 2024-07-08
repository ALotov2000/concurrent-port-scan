package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

var lastID = 0

const MinPortNum uint = 0
const MaxPortNum uint = 65535

var timeouts = []time.Duration{
	200 * time.Millisecond,
	500 * time.Millisecond,
	10 * time.Second,
}

type myTask struct {
	id            int
	ipAddr        string
	startPort     uint
	endPort       uint // inclusive
	outputChannel chan Output
	wg            *sync.WaitGroup
}

func NewTask(ipAddr string, startPort uint, endPort uint, outputChannel chan Output, wg *sync.WaitGroup) Task {
	lastID += 1
	log.Printf("myTask with id %d is created: Address=%s, ports=[%d, %d)", lastID, ipAddr, startPort, endPort)

	return &myTask{
		id:            lastID,
		ipAddr:        ipAddr,
		startPort:     startPort,
		endPort:       endPort, // inclusive
		outputChannel: outputChannel,
		wg:            wg,
	}
}

func (t *myTask) Handle() {
	defer t.wg.Done()
	log.Printf("Handling task %d. Address: %s, port Range: [%d, %d)", t.id, t.ipAddr, t.startPort, t.endPort)
	var wg sync.WaitGroup
	for port := t.startPort; port < t.endPort; port += 1 {
		address := fmt.Sprintf("%s:%d", t.ipAddr, port)
		wg.Add(1)
		go tryPort(address, port, &wg, t.outputChannel)
	}
	wg.Wait()
}

func tryPort(address string, port uint, wg *sync.WaitGroup, outputChannel chan Output) {
	defer wg.Done()
	isFound := make(chan bool, 1)
	var localWg sync.WaitGroup
	for _, timeout := range timeouts {
		localWg.Add(1)
		go func() {
			defer localWg.Done()
			conn, err := net.DialTimeout("tcp", address, timeout)
			if err == nil {
				_ = conn.Close()
				isFound <- true
				return
			}
		}()
	}
	go func() {
		localWg.Wait()
		close(isFound)
	}()
	if <-isFound {
		outputChannel <- port
	}
}

func (t *myTask) ID() int {
	return t.id
}

func (t *myTask) OutputChannel() chan Output {
	return t.outputChannel
}

func (t *myTask) WaitGroup() *sync.WaitGroup {
	return t.wg
}
