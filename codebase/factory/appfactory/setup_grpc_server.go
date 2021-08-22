package appfactory

import (
	grpcserver "pkg.agungdp.dev/candi/codebase/app/grpc_server"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/config/env"
)

func setupGRPCServer(service factory.ServiceFactory) factory.AppServerFactory {
	return grpcserver.NewServer(
		service,
		grpcserver.SetTCPPort(env.BaseEnv().GRPCPort),
		grpcserver.SetSharedListener(service.GetConfig().SharedListener),
		grpcserver.SetDebugMode(env.BaseEnv().DebugMode),
		grpcserver.SetJaegerMaxPacketSize(env.BaseEnv().JaegerMaxPacketSize),
	)
}
