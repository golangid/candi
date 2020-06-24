package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"google.golang.org/grpc"
)

type grpcServer struct {
	serverEngine *grpc.Server
	service      factory.ServiceFactory
}

// NewServer create new GRPC server
func NewServer(service factory.ServiceFactory) factory.AppServerFactory {

	return &grpcServer{
		serverEngine: grpc.NewServer(
			grpc.MaxSendMsgSize(200*int(helper.MByte)), grpc.MaxRecvMsgSize(200*int(helper.MByte)),
			grpc.UnaryInterceptor(service.GetDependency().GetMiddleware().GRPCBasicAuth),
			grpc.StreamInterceptor(service.GetDependency().GetMiddleware().GRPCBasicAuthStream),
		),
		service: service,
	}
}

func (s *grpcServer) Serve() {
	grpcPort := fmt.Sprintf(":%d", config.BaseEnv().GRPCPort)
	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		panic(err)
	}

	fmt.Printf("⇨ GRPC server run at port [::]%s\n\n", grpcPort)

	// register all module
	for _, m := range s.service.GetModules() {
		if h := m.GRPCHandler(); h != nil {
			h.Register(s.serverEngine)
		}
	}

	err = s.serverEngine.Serve(listener)
	if err != nil {
		log.Println("Unexpected Error", err)
	}
}

func (s *grpcServer) Shutdown(ctx context.Context) {
	log.Println("Stopping GRPC server...")
	s.serverEngine.GracefulStop()
}
