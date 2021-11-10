package appfactory

import (
	"net/http"

	"github.com/golangid/candi/candishared"
	restserver "github.com/golangid/candi/codebase/app/rest_server"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupRESTServer setup cron worker with default config
func SetupRESTServer(service factory.ServiceFactory) factory.AppServerFactory {
	return restserver.NewServer(
		service,
		restserver.SetHTTPPort(env.BaseEnv().HTTPPort),
		restserver.SetRootPath(env.BaseEnv().HTTPRootPath),
		restserver.SetIncludeGraphQL(env.BaseEnv().UseGraphQL),
		restserver.SetRootHTTPHandler(http.HandlerFunc(candishared.HTTPRoot(string(service.Name()), env.BaseEnv().BuildNumber))),
		restserver.SetSharedListener(service.GetConfig().SharedListener),
		restserver.SetDebugMode(env.BaseEnv().DebugMode),
		restserver.SetJaegerMaxPacketSize(env.BaseEnv().JaegerMaxPacketSize),
	)
}
