package main

import (
	"log"
	"sync"
)

func worker(id int, tasks <-chan Task, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		log.Printf("Worker %d processing task %d\n", id, task.ID())
		task.Handle()
	}
}
