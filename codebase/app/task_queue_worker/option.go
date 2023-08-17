package taskqueueworker

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/golangid/candi/codebase/interfaces"
)

type (
	option struct {
		queue                    QueueStorage
		persistent               Persistent
		secondaryPersistent      Persistent
		maxClientSubscriber      int
		autoRemoveClientInterval time.Duration
		dashboardBanner          string
		dashboardPort            uint16
		dashboardAuthKey         string
		debugMode                bool
		locker                   interfaces.Locker
	}

	// OptionFunc type
	OptionFunc func(*option)
)

// SetQueue option func
func SetQueue(q QueueStorage) OptionFunc {
	return func(o *option) {
		o.queue = q
	}
}

// SetPersistent option func
func SetPersistent(p Persistent) OptionFunc {
	return func(o *option) {
		o.persistent = p
	}
}

// SetSecondaryPersistent option func
func SetSecondaryPersistent(p Persistent) OptionFunc {
	return func(o *option) {
		o.secondaryPersistent = p
	}
}

// SetMaxClientSubscriber option func
func SetMaxClientSubscriber(max int) OptionFunc {
	return func(o *option) {
		o.maxClientSubscriber = max
	}
}

// SetAutoRemoveClientInterval option func
func SetAutoRemoveClientInterval(d time.Duration) OptionFunc {
	return func(o *option) {
		o.autoRemoveClientInterval = d
	}
}

// SetDashboardBanner option func
func SetDashboardBanner(banner string) OptionFunc {
	return func(o *option) {
		o.dashboardBanner = banner
	}
}

// SetDashboardHTTPPort option func
func SetDashboardHTTPPort(port uint16) OptionFunc {
	return func(o *option) {
		o.dashboardPort = port
	}
}

// SetDebugMode option func
func SetDebugMode(debugMode bool) OptionFunc {
	return func(o *option) {
		o.debugMode = debugMode
	}
}

// SetLocker option func
func SetLocker(locker interfaces.Locker) OptionFunc {
	return func(o *option) {
		o.locker = locker
	}
}

// SetExternalWorkerHost option func, setting worker host for add job, if not empty default using http request when add job
func SetExternalWorkerHost(host string) OptionFunc {
	externalWorkerHost = host
	return func(o *option) {
		externalWorkerHost = host
	}
}

// SetDashboardBasicAuth option func
func SetDashboardBasicAuth(username, password string) OptionFunc {
	return func(o *option) {
		o.dashboardAuthKey = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	}
}

func (o *option) basicAuth(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/_next") {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/task")
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/job")
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/expired")
		}
		if o.dashboardAuthKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", `Basic realm=""`)

		auth := r.Header.Get("Authorization")
		const prefix = "Basic "
		if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid authorization"))
			return
		}

		if auth[len(prefix):] != o.dashboardAuthKey {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid authorization"))
			return
		}

		next.ServeHTTP(w, r)
	}
}
