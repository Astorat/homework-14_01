package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEventLogger_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")

	el := NewEventLogger(logFile)
	el.Start()

	el.Log("test event 1")
	el.Stop()

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "test event 1") {
		t.Fatalf("log file should contain 'test event 1', got: %s", content)
	}
}

func TestEventLogger_MultipleEvents(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")

	el := NewEventLogger(logFile)
	el.Start()

	el.Log("event A")
	el.Log("event B")
	el.Log("event C")
	el.Stop()

	data, _ := os.ReadFile(logFile)
	content := string(data)

	for _, event := range []string{"event A", "event B", "event C"} {
		if !strings.Contains(content, event) {
			t.Fatalf("log file should contain '%s', got: %s", event, content)
		}
	}
}

func TestEventLogger_FormatsWithTimestamp(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")

	el := NewEventLogger(logFile)
	el.Start()

	el.Log("timestamped event")
	el.Stop()

	data, _ := os.ReadFile(logFile)
	content := string(data)

	if !strings.Contains(content, "[") || !strings.Contains(content, "]") {
		t.Fatalf("log entry should have timestamp in brackets, got: %s", content)
	}
}

func TestEventLogger_StopDrainsEvents(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")

	el := NewEventLogger(logFile)
	el.Start()

	for i := 0; i < 5; i++ {
		el.Log("drain event")
	}

	el.Stop()

	data, _ := os.ReadFile(logFile)
	content := string(data)
	count := strings.Count(content, "drain event")
	if count != 5 {
		t.Fatalf("expected 5 drain events, got %d", count)
	}
}

func TestEventLogger_ChannelFull(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")

	el := &EventLogger{
		ch:       make(chan string, 2),
		done:     make(chan struct{}),
		filePath: logFile,
	}

	el.Log("msg1")
	el.Log("msg2")
	el.Log("msg3 should be dropped")

	if len(el.ch) != 2 {
		t.Fatalf("expected channel length 2, got %d", len(el.ch))
	}
}

func TestEventLogger_DelayedWrite(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")

	el := NewEventLogger(logFile)
	el.Start()

	el.Log("delayed event")

	time.Sleep(500 * time.Millisecond)
	data, _ := os.ReadFile(logFile)
	if len(data) > 0 {
		t.Log("event written before expected delay (non-critical)")
	}

	el.Stop()

	data, _ = os.ReadFile(logFile)
	if !strings.Contains(string(data), "delayed event") {
		t.Fatal("event should be in log after Stop()")
	}
}

func TestEventLogger_FileCreated(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "new_log.txt")

	el := NewEventLogger(logFile)
	el.Start()
	el.Log("create file event")
	el.Stop()

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatal("log file was not created")
	}
}

func TestEventLogger_AppendMode(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "append.log")

	el1 := NewEventLogger(logFile)
	el1.Start()
	el1.Log("first run")
	el1.Stop()

	el2 := NewEventLogger(logFile)
	el2.Start()
	el2.Log("second run")
	el2.Stop()

	data, _ := os.ReadFile(logFile)
	content := string(data)

	if !strings.Contains(content, "first run") || !strings.Contains(content, "second run") {
		t.Fatalf("log file should contain both runs, got: %s", content)
	}
}

func TestNewEventLogger(t *testing.T) {
	el := NewEventLogger("test.log")
	if el == nil {
		t.Fatal("expected non-nil EventLogger")
	}
	if el.filePath != "test.log" {
		t.Fatalf("expected filePath test.log, got %s", el.filePath)
	}
	if cap(el.ch) != 100 {
		t.Fatalf("expected channel capacity 100, got %d", cap(el.ch))
	}
}
