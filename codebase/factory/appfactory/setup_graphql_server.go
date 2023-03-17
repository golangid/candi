package appfactory

import (
	graphqlserver "github.com/golangid/candi/codebase/app/graphql_server"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupGraphQLServer setup graphql server with default config
func SetupGraphQLServer(service factory.ServiceFactory, opts ...graphqlserver.OptionFunc) factory.AppServerFactory {
	gqlOptions := []graphqlserver.OptionFunc{
		graphqlserver.SetHTTPPort(env.BaseEnv().HTTPPort),
		graphqlserver.SetRootPath(env.BaseEnv().HTTPRootPath),
		graphqlserver.SetDisableIntrospection(env.BaseEnv().GraphQLDisableIntrospection),
		graphqlserver.SetSharedListener(service.GetConfig().SharedListener),
		graphqlserver.SetDebugMode(env.BaseEnv().DebugMode),
		graphqlserver.SetJaegerMaxPacketSize(env.BaseEnv().JaegerMaxPacketSize),
	}
	gqlOptions = append(gqlOptions, opts...)
	return graphqlserver.NewServer(service, gqlOptions...)
}
