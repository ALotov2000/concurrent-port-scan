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
const Timeout = 5 * time.Second
const NumRepeat = 2

type myTask struct {
	id            int
	ipAddr        string
	startPort     uint
	endPort       uint // inclusive
	outputChannel chan Output
}

func NewTask(ipAddr string, startPort uint, endPort uint, outputChannel chan Output) Task {
	lastID += 1
	log.Printf("myTask with id %d is created: Address=%s, ports=[%d, %d)", lastID, ipAddr, startPort, endPort)

	return &myTask{
		id:            lastID,
		ipAddr:        ipAddr,
		startPort:     startPort,
		endPort:       endPort, // inclusive
		outputChannel: outputChannel,
	}
}

func (t *myTask) Handle() {
	log.Printf("Handling task %d. Address: %s, port Range: [%d, %d)", t.id, t.ipAddr, t.startPort, t.endPort)
	var wg sync.WaitGroup
	for port := t.startPort; port < t.endPort; port += 1 {
		wg.Add(1)
		port := port
		go func() {
			defer wg.Done()
			address := fmt.Sprintf("%s:%d", t.ipAddr, port)
			for range NumRepeat {
				conn, err := net.DialTimeout("tcp", address, Timeout)
				if err == nil {
					_ = conn.Close()
					t.outputChannel <- port
					return
				}
			}
		}()
	}
	wg.Wait()
}

func (t *myTask) ID() int {
	return t.id
}

func (t *myTask) OutputChannel() chan Output {
	return t.outputChannel
}
