package taskqueueworker

const schema = `schema {
	query: Query
	mutation: Mutation
	subscription: Subscription
}

type Query {
	tagline(): TaglineType!
}

type Mutation {
	add_job(task_name: String!, max_retry: Int!, args: String!): String!
	stop_job(job_id: String!): String!
	stop_all_job(task_name: String!): String!
	retry_job(job_id: String!): String!
	clean_job(task_name: String!): String!
	retry_all_job(task_name: String!): String!
}

type Subscription {
	listen_task(): TaskListResolver!
	listen_task_job_detail(task_name: String!, page: Int!, limit: Int!, search: String, status: [String!]!): JobListResolver!
}

type TaglineType {
	version: String!
	banner: String!
	tagline: String!
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
}

type MemstatsResolver {
	alloc: String!
	total_alloc: String!
	num_gc: Int!
	num_goroutines: Int!
}

type TaskResolver {
	name: String!
	total_jobs: Int!
	detail: TaskDetailResolver!
}

type TaskListResolver {
	meta: MetaTaskResolver!
	data: [TaskResolver!]!
}

type TaskDetailResolver {
	give_up: Int!
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
	status: String!
	created_at: String!
	finished_at: String!
	next_retry_at: String!
}`
