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
}

type Subscription {
	subscribe_all_task(): [TaskType!]!
	listen_task(task_name: String!, page: Int!, limit: Int!, search: String, status: [String!]!): JobListType!
}

type TaglineType {
	tagline: String!
	task_list_client_subscribers: [String!]!
	job_list_client_subscribers: [String!]!
}

type MetaType {
	page: Int!
	limit: Int!
	total_pages: Int!
	total_records: Int!
	detail: TaskDetailType!
}

type TaskType {
	name: String!
	total_jobs: Int!
	detail: TaskDetailType!
}

type TaskDetailType {
	give_up: Int!
	retrying: Int!
	success: Int!
	queueing: Int!
	stopped: Int!
}

type JobListType {
	meta: MetaType!
	data: [JobType!]!
}

type JobType {
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
