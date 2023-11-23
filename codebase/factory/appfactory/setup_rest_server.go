package appfactory

import (
	graphqlserver "github.com/golangid/candi/codebase/app/graphql_server"
	restserver "github.com/golangid/candi/codebase/app/rest_server"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupRESTServer setup rest server with default config
func SetupRESTServer(service factory.ServiceFactory, opts ...restserver.OptionFunc) factory.AppServerFactory {
	restOptions := []restserver.OptionFunc{
		restserver.SetHTTPPort(env.BaseEnv().HTTPPort),
		restserver.SetRootPath(env.BaseEnv().HTTPRootPath),
		restserver.SetIncludeGraphQL(env.BaseEnv().UseGraphQL),
		restserver.SetSharedListener(service.GetConfig().SharedListener),
		restserver.SetDebugMode(env.BaseEnv().DebugMode),
		restserver.SetJaegerMaxPacketSize(env.BaseEnv().JaegerMaxPacketSize),
	}
	if env.BaseEnv().UseGraphQL {
		restOptions = append(restOptions, restserver.AddGraphQLOption(
			graphqlserver.SetDisableIntrospection(env.BaseEnv().GraphQLDisableIntrospection),
			graphqlserver.SetHTTPPort(env.BaseEnv().HTTPPort),
		))
	}
	restOptions = append(restOptions, opts...)
	return restserver.NewServer(service, restOptions...)
}
