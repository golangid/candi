package appfactory

import (
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

/*
NewAppFromEnvironmentConfig constructor

Construct server/worker for running application from environment value

## Server

USE_REST=[bool]

USE_GRPC=[bool]

USE_GRAPHQL=[bool]

## Worker

USE_KAFKA_CONSUMER=[bool] # event driven handler

USE_CRON_SCHEDULER=[bool] # static scheduler

USE_REDIS_SUBSCRIBER=[bool] # dynamic scheduler

USE_TASK_QUEUE_WORKER=[bool]

USE_POSTGRES_LISTENER_WORKER=[bool]

USE_RABBITMQ_CONSUMER=[bool] # event driven handler and dynamic scheduler
*/
func NewAppFromEnvironmentConfig(service factory.ServiceFactory) (apps []factory.AppServerFactory) {

	if env.BaseEnv().UseKafkaConsumer {
		apps = append(apps, SetupKafkaWorker(service))
	}
	if env.BaseEnv().UseCronScheduler {
		apps = append(apps, SetupCronWorker(service))
	}
	if env.BaseEnv().UseTaskQueueWorker {
		apps = append(apps, SetupTaskQueueWorker(service))
	}
	if env.BaseEnv().UseRedisSubscriber {
		apps = append(apps, SetupRedisWorker(service))
	}
	if env.BaseEnv().UsePostgresListenerWorker {
		apps = append(apps, SetupPostgresWorker(service))
	}
	if env.BaseEnv().UseRabbitMQWorker {
		apps = append(apps, SetupRabbitMQWorker(service))
	}

	if env.BaseEnv().UseREST {
		apps = append(apps, SetupRESTServer(service))
	}
	if env.BaseEnv().UseGRPC {
		apps = append(apps, SetupGRPCServer(service))
	}
	if !env.BaseEnv().UseREST && env.BaseEnv().UseGraphQL {
		apps = append(apps, SetupGraphQLServer(service))
	}

	return
}
