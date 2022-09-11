package appfactory

import (
	grpcserver "github.com/golangid/candi/codebase/app/grpc_server"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupGRPCServer setup grpc server with default config
func SetupGRPCServer(service factory.ServiceFactory, opts ...grpcserver.OptionFunc) factory.AppServerFactory {
	grpcOption := []grpcserver.OptionFunc{
		grpcserver.SetTCPPort(env.BaseEnv().GRPCPort),
		grpcserver.SetSharedListener(service.GetConfig().SharedListener),
		grpcserver.SetDebugMode(env.BaseEnv().DebugMode),
		grpcserver.SetJaegerMaxPacketSize(env.BaseEnv().JaegerMaxPacketSize),
	}
	grpcOption = append(grpcOption, opts...)
	return grpcserver.NewServer(service, grpcOption...)
}
