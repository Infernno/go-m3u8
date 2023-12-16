package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Logger interface {
	Info(tag string, message string)
	Error(tag string, message string)

	Shutdown()
}

type FileLogger struct {
	filePath string
	queue    []string
	mutex    *sync.Mutex
	delay    time.Duration
	finished bool
}

func NewFileLogger(path string) *FileLogger {
	logger := &FileLogger{
		queue:    make([]string, 0),
		filePath: path,
		mutex:    &sync.Mutex{},
		delay:    time.Duration(1000),
		finished: false,
	}

	go logger.launchGoroutine()

	return logger
}

func (logger *FileLogger) Info(tag string, message string) {
	go logger.withLock(func() {
		logger.queue = append(logger.queue, fmt.Sprintf("[INFO][%s][%s]: %s\n", tag, time.Now().Format(time.RFC850), message))
	})
}

func (logger *FileLogger) Error(tag string, message string) {
	go logger.withLock(func() {
		logger.queue = append(logger.queue, fmt.Sprintf("[ERROR][%s][%s]: %s\n", tag, time.Now().Format(time.RFC850), message))
	})
}

func (logger *FileLogger) Shutdown() {
	logger.withLock(func() {
		logger.finished = true
		logger.flush(true)
	})
}

type action func()

func (logger *FileLogger) launchGoroutine() {
	for {
		time.Sleep(logger.delay)

		isFinished := false

		logger.withLock(func() {
			isFinished = logger.finished

			logger.flush(false)

			if isFinished {
				return
			}
		})

		if isFinished {
			fmt.Println("Exit logger goroutine")
			return
		}
	}
}

func (logger *FileLogger) flush(sysFlush bool) {
	file, err := os.OpenFile(logger.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)

	if err != nil {
		fmt.Println("ERROR - ", err)
		return
	}

	defer file.Close()

	for _, t := range logger.queue {
		_, wErr := file.WriteString(t)

		if wErr != nil {
			fmt.Println("Failed to write string to file", logger.filePath, "because of", wErr)
		}

		logger.queue = logger.queue[1:]
	}

	if sysFlush {
		err := file.Sync()

		if err != nil {
			fmt.Println("Sync failed:", err)
		}
	}
}

func (logger *FileLogger) withLock(action action) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	action()
}
