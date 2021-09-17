package appfactory

import (
	grpcserver "github.com/golangid/candi/codebase/app/grpc_server"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
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
