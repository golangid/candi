package taskqueueworker

func convertIncrementMap(mp map[string]int) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range mp {
		res[k] = v
	}
	return res
}

func isDefaultPersistent() bool {
	_, ok := persistent.(*noopPersistent)
	return ok
}

func isDefaultQueue() bool {
	_, ok := queue.(*inMemQueue)
	return ok
}
