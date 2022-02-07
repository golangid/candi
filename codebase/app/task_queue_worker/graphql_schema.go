package taskqueueworker

const schema = `schema {
	query: Query
	mutation: Mutation
	subscription: Subscription
}

type Query {
	tagline(): TaglineType!
	get_job_detail(job_id: String!): JobResolver
}

type Mutation {
	add_job(task_name: String!, max_retry: Int!, args: String!): String!
	stop_job(job_id: String!): String!
	stop_all_job(task_name: String!): String!
	retry_job(job_id: String!): String!
	clean_job(task_name: String!): String!
	retry_all_job(task_name: String!): String!
	clear_all_client_subscriber(): String!
	delete_job(job_id: String!): String!
}

type Subscription {
	listen_task(): TaskListResolver!
	listen_task_job_detail(task_name: String!, page: Int!, limit: Int!, search: String, status: [String!]!): JobListResolver!
}

type TaglineType {
	version: String!
	banner: String!
	tagline: String!
	start_at: String!
	build_number: String!
	task_list_client_subscribers: [String!]!
	job_list_client_subscribers: [String!]!
	memory_statistics: MemstatsResolver!
}

type MetaType {
	page: Int!
	limit: Int!
	total_pages: Int!
	total_records: Int!
	is_close_session: Boolean!
	detail: TaskDetailResolver!
}

type MetaTaskResolver {
	page: Int!
	limit: Int!
	total_pages: Int!
	total_records: Int!
	is_close_session: Boolean!
	total_client_subscriber: Int!
}

type MemstatsResolver {
	alloc: String!
	total_alloc: String!
	num_gc: Int!
	num_goroutines: Int!
}

type TaskResolver {
	name: String!
	module_name: String!
	total_jobs: Int!
	detail: TaskDetailResolver!
}

type TaskListResolver {
	meta: MetaTaskResolver!
	data: [TaskResolver!]!
}

type TaskDetailResolver {
	failure: Int!
	retrying: Int!
	success: Int!
	queueing: Int!
	stopped: Int!
}

type JobListResolver {
	meta: MetaType!
	data: [JobResolver!]!
}

type JobResolver {
	id: String!
	task_name: String!
	arguments: String!
	retries: Int!
	max_retry: Int!
	interval: String!
	error: String!
	trace_id: String!
	retry_histories: [JobRetryHistory!]!
	status: String!
	created_at: String!
	finished_at: String!
	next_retry_at: String!
}

type JobRetryHistory {
	error_stack: String!
	status: String!
	error: String!
	trace_id: String!
	start_at: String!
	end_at: String!
}
`
