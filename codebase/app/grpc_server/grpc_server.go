package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/config"
	"pkg.agungdwiprasetyo.com/candi/logger"
)

type grpcServer struct {
	serverEngine *grpc.Server
	service      factory.ServiceFactory
}

// NewServer create new GRPC server
func NewServer(service factory.ServiceFactory) factory.AppServerFactory {

	return &grpcServer{
		serverEngine: grpc.NewServer(
			grpc.MaxSendMsgSize(200*int(candihelper.MByte)), grpc.MaxRecvMsgSize(200*int(candihelper.MByte)),
			grpc.UnaryInterceptor(chainUnaryServer(
				unaryTracerInterceptor,
				unaryLogInterceptor,
				unaryPanicInterceptor,
			)),
			grpc.StreamInterceptor(chainStreamServer(
				streamTracerInterceptor,
				streamLogInterceptor,
				streamPanicInterceptor,
			)),
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

	fmt.Printf("\x1b[34;1m⇨ GRPC server run at port [::]%s\x1b[0m\n\n", grpcPort)

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
	deferFunc := logger.LogWithDefer("Stopping GRPC server...")
	defer deferFunc()

	s.serverEngine.GracefulStop()
}
