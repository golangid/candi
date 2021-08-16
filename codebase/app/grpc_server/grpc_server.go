package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
)

type grpcServer struct {
	serverEngine *grpc.Server
	listener     net.Listener
	service      factory.ServiceFactory
}

// NewServer create new GRPC server
func NewServer(service factory.ServiceFactory, muxListener cmux.CMux) factory.AppServerFactory {

	var (
		kaep = keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
			PermitWithoutStream: true,            // Allow pings even when there are no active streams
		}
		kasp = keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
			MaxConnectionAgeGrace: 10 * time.Second, // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
			Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
			Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
		}
	)

	intercept := new(interceptor)

	server := &grpcServer{
		serverEngine: grpc.NewServer(
			grpc.KeepaliveEnforcementPolicy(kaep),
			grpc.KeepaliveParams(kasp),
			grpc.MaxSendMsgSize(200*int(candihelper.MByte)), grpc.MaxRecvMsgSize(200*int(candihelper.MByte)),
			grpc.UnaryInterceptor(chainUnaryServer(
				intercept.unaryTracerInterceptor,
				intercept.unaryMiddlewareInterceptor,
			)),
			grpc.StreamInterceptor(chainStreamServer(
				intercept.streamTracerInterceptor,
				intercept.streamMiddlewareInterceptor,
			)),
		),
		service: service,
	}

	grpcPort := fmt.Sprintf(":%d", env.BaseEnv().GRPCPort)
	if muxListener == nil {
		var err error
		server.listener, err = net.Listen("tcp", grpcPort)
		if err != nil {
			panic(err)
		}
	} else {
		server.listener = muxListener.MatchWithWriters(
			cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
			cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc+proto"),
		)
		grpcPort = server.listener.Addr().String()
	}

	// register all module
	intercept.middleware = make(types.MiddlewareGroup)
	for _, m := range service.GetModules() {
		if h := m.GRPCHandler(); h != nil {
			h.Register(server.serverEngine, &intercept.middleware)
		}
	}

	for root, info := range server.serverEngine.GetServiceInfo() {
		for _, method := range info.Methods {
			logger.LogGreen(fmt.Sprintf("[GRPC-METHOD] /%s/%s \t\t[metadata]--> %v", root, method.Name, info.Metadata))
		}
	}
	fmt.Printf("\x1b[34;1m⇨ GRPC server run at port [::]%s\x1b[0m\n\n", grpcPort)

	return server
}

func (s *grpcServer) Serve() {
	if err := s.serverEngine.Serve(s.listener); err != nil {
		log.Println("GRPC: Unexpected Error", err)
	}
}

func (s *grpcServer) Shutdown(ctx context.Context) {
	defer log.Println("\x1b[33;1mStopping GRPC server:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")

	s.serverEngine.GracefulStop()
	s.listener.Close()
}

func (s *grpcServer) Name() string {
	return string(types.GRPC)
}
