package taskqueueworker

import "errors"

var (
	errClientLimitExceeded = errors.New("client limit exceeded, please try again later")
)

func convertIncrementMap(mp map[string]int) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range mp {
		res[k] = v
	}
	return res
}

func isDefaultPersistent() bool {
	_, ok := engine.opt.persistent.(*noopPersistent)
	return ok
}

func isDefaultQueue() bool {
	_, ok := engine.opt.queue.(*inMemQueue)
	return ok
}
