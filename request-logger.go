package githubwebhookdeploy

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type rLog struct {
	time    time.Time
	message string
}
type requestLog struct {
	mux   sync.Mutex
	logs  []rLog
	guids map[string]bool
}

func NewRequestLog() *requestLog {
	return &requestLog{guids: make(map[string]bool)}
}

var requestLogger = NewRequestLog()

func (r *requestLog) AddGUID(guid string) {
	r.mux.Lock()
	r.guids[guid] = true
	r.mux.Unlock()
}

func (r *requestLog) GUIDExists(guid string) bool {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.guids[guid]
}

// Logf stores the log messages & logs them to stdout
func (r *requestLog) Logf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	now := time.Now()
	log.Print(msg)
	r.mux.Lock()
	r.logs = append(r.logs, rLog{time: now, message: msg})
	r.mux.Unlock()
}

// GetLogs returns the stored log messages
func (r *requestLog) GetLogs() []string {
	r.mux.Lock()
	defer r.mux.Unlock()
	var allLogs []string
	for _, l := range r.logs {
		allLogs = append(allLogs, fmt.Sprintf("%s: %s", l.time.Format("2006-01-02 15:04:05"), l.message))
	}
	return allLogs
}
