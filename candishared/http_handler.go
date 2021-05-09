package candishared

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"pkg.agungdp.dev/candi/config/env"
)

// HTTPRoot http handler
func HTTPRoot(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]string{
			"message":   fmt.Sprintf("Service %s up and running", serviceName),
			"timestamp": time.Now().Format(time.RFC3339Nano),
		}
		if env.BaseEnv().BuildNumber != "" {
			payload["build_number"] = env.BaseEnv().BuildNumber
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}
}

// HTTPMemstatsHandler calculate runtime statistic
func HTTPMemstatsHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	data := struct {
		NumGoroutine int         `json:"num_goroutine"`
		Memstats     interface{} `json:"memstats"`
	}{
		runtime.NumGoroutine(), m,
	}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
