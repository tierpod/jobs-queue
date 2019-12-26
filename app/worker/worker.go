// Package worker describes Worker. Each Worker takes Job from channel and executes it. Logs all
// stdout and stderr messages.
package worker

import (
	"bufio"
	"bytes"
	"log"

	"github.com/tierpod/jobs-queue/app/cache"
)

// Worker includes queue and identifier.
type Worker struct {
	queue <-chan Command
	cache *cache.Cache
	name  string
}

// New creates new Worker. Logs messages from this logger with prefix `name`.
func New(inCh <-chan Command, inCache *cache.Cache, name string) *Worker {
	return &Worker{
		cache: inCache,
		queue: inCh,
		name:  name,
	}
}

// Start starts Worker. If you need to start Worker in background, use `go` keyword.
func (w Worker) Start() {
	log.Printf("[INFO]  (%v) start", w.name)
	for {
		select {
		case cmd := <-w.queue:
			w.process(cmd)
		}
	}
}

func (w Worker) process(cmd Command) {
	w.cache.Set(cmd.Key(), nil)
	defer w.cache.Del(cmd.Key())

	log.Printf("[INFO]  (%s) exec  : %v", w.name, cmd)
	stdout, stderr, err := cmd.Exec()
	if err != nil {
		log.Printf("[ERROR] (%s) %v: %v", w.name, cmd, err)
		return
	}

	if stdout.Len() > 0 {
		buf := bytes.NewReader(stdout.Bytes())
		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			log.Printf("[INFO]  (%s) stdout: %v", w.name, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Printf("[ERROR] (%s) stdout: %v", w.name, err)
		}
	}

	if stderr.Len() > 0 {
		buf := bytes.NewReader(stderr.Bytes())
		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			log.Printf("[INFO]  (%s) stderr: %v", w.name, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Printf("[ERROR] (%s) stderr: %v", w.name, err)
		}
	}
}
