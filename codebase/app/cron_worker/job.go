package cronworker

import (
	"time"

	cronexpr "github.com/golangid/candi/candiutils/cronparser"
	"github.com/golangid/candi/codebase/factory/types"
)

// Job model
type Job struct {
	HandlerName  string              `json:"handler_name"`
	Interval     string              `json:"interval"`
	Handler      types.WorkerHandler `json:"-"`
	Params       string              `json:"params"`
	WorkerIndex  int                 `json:"worker_index"`
	ticker       *time.Ticker        `json:"-"`
	schedule     cronexpr.Schedule   `json:"-"`
	nextDuration *time.Duration      `json:"-"`
}
