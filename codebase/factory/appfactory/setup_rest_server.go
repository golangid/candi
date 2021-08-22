package appfactory

import (
	"net/http"

	"pkg.agungdp.dev/candi/candishared"
	restserver "pkg.agungdp.dev/candi/codebase/app/rest_server"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/config/env"
)

func setupRESTServer(service factory.ServiceFactory) factory.AppServerFactory {
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
