package wrapper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/golangid/candi/config/env"
)

// HTTPHandlerDefaultRoot default root http handler
func HTTPHandlerDefaultRoot(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	payload := struct {
		BuildNumber string `json:"build_number,omitempty"`
		Message     string `json:"message,omitempty"`
		Hostname    string `json:"hostname,omitempty"`
		Timestamp   string `json:"timestamp,omitempty"`
		StartAt     string `json:"start_at,omitempty"`
		Uptime      string `json:"uptime,omitempty"`
	}{
		Message:   fmt.Sprintf("Service %s up and running", env.BaseEnv().ServiceName),
		Timestamp: now.Format(time.RFC3339Nano),
	}

	if startAt, err := time.Parse(time.RFC3339, env.BaseEnv().StartAt); err == nil {
		payload.StartAt = env.BaseEnv().StartAt
		payload.Uptime = now.Sub(startAt).String()
	}
	if env.BaseEnv().BuildNumber != "" {
		payload.BuildNumber = env.BaseEnv().BuildNumber
	}
	if hostname, err := os.Hostname(); err == nil {
		payload.Hostname = hostname
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

// HTTPHandlerMemstats calculate runtime statistic
func HTTPHandlerMemstats(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	data := struct {
		NumGoroutine int         `json:"num_goroutine"`
		Memstats     interface{} `json:"memstats"`
	}{
		runtime.NumGoroutine(), m,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
