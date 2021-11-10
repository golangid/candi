package appfactory

import (
	"net/http"

	"github.com/golangid/candi/candishared"
	graphqlserver "github.com/golangid/candi/codebase/app/graphql_server"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupGraphQLServer setup cron worker with default config
func SetupGraphQLServer(service factory.ServiceFactory) factory.AppServerFactory {
	return graphqlserver.NewServer(
		service,
		graphqlserver.SetHTTPPort(env.BaseEnv().HTTPPort),
		graphqlserver.SetRootPath(env.BaseEnv().HTTPRootPath),
		graphqlserver.SetDisableIntrospection(env.BaseEnv().GraphQLDisableIntrospection),
		graphqlserver.SetRootHTTPHandler(http.HandlerFunc(candishared.HTTPRoot(string(service.Name()), env.BaseEnv().BuildNumber))),
		graphqlserver.SetSharedListener(service.GetConfig().SharedListener),
		graphqlserver.SetDebugMode(env.BaseEnv().DebugMode),
		graphqlserver.SetJaegerMaxPacketSize(env.BaseEnv().JaegerMaxPacketSize),
	)
}
