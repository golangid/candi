package taskqueueworker

import "errors"

var (
	errClientLimitExceeded = errors.New("client limit exceeded, please try again later")
	errWorkerInactive      = errors.New("Worker is inactive")
)

func convertIncrementMap(mp map[string]int) map[string]any {
	res := make(map[string]any)
	for k, v := range mp {
		res[k] = v
	}
	return res
}

func normalizeCount(count int) int {
	if count < 0 {
		return 0
	}
	return count
}
