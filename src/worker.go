package main

import (
	"context"
	"log"
	"sync"
)

func worker(id uint, tasks <-chan Task, wg *sync.WaitGroup, ctx *context.Context) {
	defer wg.Done()
	for {
		select {
		case <-(*ctx).Done():
			log.Printf("Worker %d received shutdown signal\n", id)
			return
		case task, ok := <-tasks:
			if !ok {
				log.Printf("Worker %d found task channel closed\n", id)
				return
			}
			log.Printf("Worker %d processing task %d\n", id, task.ID())
			task.Handle()
		}
	}
}
