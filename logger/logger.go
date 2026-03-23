package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type EventLogger struct {
	ch       chan string
	done     chan struct{}
	filePath string
}

func NewEventLogger(filePath string) *EventLogger {
	return &EventLogger{
		ch:       make(chan string, 100),
		done:     make(chan struct{}),
		filePath: filePath,
	}
}

// Start запускает горутину-воркер, которая читает события из канала
// и записывает их в файл с задержкой 1 секунда.
func (el *EventLogger) Start() {
	go func() {
		file, err := os.OpenFile(el.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Failed to open log file: %v", err)
			close(el.done)
			return
		}
		defer file.Close()
		defer close(el.done)

		for event := range el.ch {
			time.Sleep(1 * time.Second)
			entry := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), event)
			if _, err := file.WriteString(entry); err != nil {
				log.Printf("Failed to write log entry: %v", err)
			}
			log.Printf("Logged event: %s", event)
		}
	}()
}

// Log отправляет событие в канал для отложенной записи.
func (el *EventLogger) Log(event string) {
	select {
	case el.ch <- event:
	default:
		log.Printf("Logger channel full, dropping event: %s", event)
	}
}

// Stop закрывает канал и ожидает завершения горутины-воркера.
func (el *EventLogger) Stop() {
	close(el.ch)
	<-el.done
}
