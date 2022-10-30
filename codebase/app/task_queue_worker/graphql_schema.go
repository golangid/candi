package taskqueueworker

const schema = `schema {
	query: Query
	mutation: Mutation
	subscription: Subscription
}

type Query {
	dashboard(gc: Boolean): DashboardType!
	get_detail_job(job_id: String!, filter: GetAllJobHistoryInputResolver): JobResolver!
	get_all_active_subscriber(): [ClientSubscriber!]!
	get_all_job(filter: GetAllJobInputResolver): JobListResolver!
	get_all_configuration(): [ConfigurationResolver!]!
}

type Mutation {
	add_job(param: AddJobInputResolver!): String!
	stop_job(job_id: String!): String!
	stop_all_job(task_name: String!): String!
	retry_job(job_id: String!): String!
	clean_job(task_name: String!): String!
	retry_all_job(task_name: String!): String!
	clear_all_client_subscriber(): String!
	kill_client_subscriber(client_id: String!): String!
	delete_job(job_id: String!): String!
	recalculate_summary(): String!
	set_configuration(config: SetConfigurationInputResolver!): String!
	run_queued_job(task_name: String!): String!
	restore_from_secondary(): RestoreSecondaryResolver!
}

type Subscription {
	listen_task_dashboard(
		page: Int!,
		limit: Int!,
		search: String
	): TaskListResolver!
	listen_all_job(filter: GetAllJobInputResolver): JobListResolver!
	listen_detail_job(job_id: String!, filter: GetAllJobHistoryInputResolver): JobResolver!
}

type DashboardType {
	version: String!
	go_version: String!
	banner: String!
	tagline: String!
	start_at: String!
	build_number: String!
	config: Config!
	memory_statistics: MemstatsResolver!
	dependency_health: DependencyHealth!
	dependency_detail: DependencyDetail!
}

type Config {
	with_persistent: Boolean!
}

type DependencyHealth {
	persistent: String
	queue: String
}

type DependencyDetail {
	persistent_type: String!
	queue_type: String!
	use_secondary_persistent: Boolean!
}

type MetaType {
	page: Int!
	limit: Int!
	total_pages: Int!
	total_records: Int!
	is_close_session: Boolean!
	is_loading: Boolean!
	is_freeze_broadcast: Boolean!
	detail: TaskDetailResolver!
}

type MetaTaskResolver {
	page: Int!
	limit: Int!
	total_pages: Int!
	total_records: Int!
	is_close_session: Boolean!
	total_client_subscriber: Int!
	client_id: String!
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
	is_loading: Boolean!
	loading_message: String!
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
	meta: JoDetailMetaResolver!
}

type JoDetailMetaResolver {
	is_close_session: Boolean!
	page: Int!
	total_history: Int!
}

type JobRetryHistory {
	error_stack: String!
	status: String!
	error: String!
	trace_id: String!
	start_at: String!
	end_at: String!
}

type FilterJobList {
	task_name: String!
	page: Int!
	limit: Int!
	search: String
	statuses: [String!]!
	start_date: String!
	end_date: String!
}

type ClientSubscriber {
	client_id: String!
	page_name: String!
	page_filter: String!
}

input AddJobInputResolver {
	task_name: String!
	max_retry: Int!
	args: String!
	retry_interval: String
}

input GetAllJobInputResolver {
	page: Int, 
	limit: Int, 
	task_name: String, 
	search: String, 
	statuses: [String!],
	start_date: String,
	end_date: String,
	job_id: String
}

input GetAllJobHistoryInputResolver {
	page: Int, 
	limit: Int,
	start_date: String,
	end_date: String
}

input SetConfigurationInputResolver {
	key: String!
	name: String!
	value: String!
	is_active: Boolean!
}

type ConfigurationResolver {
	key: String!
	name: String!
	value: String!
	is_active: Boolean!
}

type RestoreSecondaryResolver {
	total_data: Int!
	message: String!
}
`

// AddJobInputResolver model
type AddJobInputResolver struct {
	TaskName      string
	MaxRetry      int
	Args          string
	RetryInterval *string
}
