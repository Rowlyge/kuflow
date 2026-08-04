package telemetry

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu sync.Mutex
	f  *os.File
}

func NewLogger(path string) (*Logger, error) {

	f, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, err
	}

	return &Logger{
		f: f,
	}, nil
}

func (l *Logger) Log(v any) {

	l.mu.Lock()
	defer l.mu.Unlock()

	_ = json.NewEncoder(l.f).Encode(v)
}

func (l *Logger) Close() error {
	return l.f.Close()
}

type ConnectEvent struct {
	Timestamp time.Time `json:"timestamp"`

	Type string `json:"type"`

	Client string `json:"client"`

	Target string `json:"target"`

	DurationMs int64 `json:"duration_ms"`

	ClientToTargetBytes int64 `json:"client_to_target_bytes"`

	TargetToClientBytes int64 `json:"target_to_client_bytes"`
}

var Default *Logger

func Init(path string) {

	logger, err := NewLogger(path)
	if err != nil {
		log.Fatal(err)
	}

	Default = logger
}
