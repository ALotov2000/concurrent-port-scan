package main

import "log"

var lastID = 0

type Task struct {
	ID        int
	IPAddr    string
	StartPort int
	EndPort   int
	Results   chan Result
}

func NewTask(ipAddr string, startPort int, endPort int) *Task {

	lastID += 1
	log.Printf("Task with id %d is created: IP=%s, ports=[%d, %d)", lastID, ipAddr, startPort, endPort)

	return &Task{
		ID:        lastID,
		IPAddr:    ipAddr,
		StartPort: startPort,
		EndPort:   endPort, // exclusive
		Results:   make(chan Result),
	}
}

func (t *Task) Handle() {
	log.Printf("Handling task %d. IP: %s, Port Range: [%d, %d)", t.ID, t.IPAddr, t.StartPort, t.EndPort)
	// todo
	close(t.Results)
}
