package appfactory

import (
	"net/http"

	"pkg.agungdp.dev/candi/candishared"
	graphqlserver "pkg.agungdp.dev/candi/codebase/app/graphql_server"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/config/env"
)

func setupGraphQLServer(service factory.ServiceFactory) factory.AppServerFactory {
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
