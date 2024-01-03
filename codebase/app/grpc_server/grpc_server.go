package grpcserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

type grpcServer struct {
	opt          option
	serverEngine *grpc.Server
	listener     net.Listener
	service      factory.ServiceFactory
}

// NewServer create new GRPC server
func NewServer(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {
	intercept := new(interceptor)
	serverOpt := getDefaultOption()
	for _, opt := range opts {
		opt(&serverOpt)
	}
	serverOpt.serverOptions = append(serverOpt.serverOptions,
		grpc.UnaryInterceptor(chainUnaryServer(
			intercept.unaryTracerInterceptor,
			intercept.unaryMiddlewareInterceptor,
		)),
		grpc.StreamInterceptor(chainStreamServer(
			intercept.streamTracerInterceptor,
			intercept.streamMiddlewareInterceptor,
		)))

	server := &grpcServer{
		serverEngine: grpc.NewServer(serverOpt.serverOptions...),
		service:      service,
		opt:          serverOpt,
	}

	grpcPort := server.opt.tcpPort
	if server.opt.sharedListener == nil {
		var err error
		server.listener, err = net.Listen("tcp", grpcPort)
		if err != nil {
			panic(err)
		}
	} else {
		server.listener = server.opt.sharedListener.MatchWithWriters(
			cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
			cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc+proto"),
		)
		grpcPort = server.listener.Addr().String()
	}

	if server.opt.tlsConfig != nil {
		server.listener = tls.NewListener(server.listener, server.opt.tlsConfig)
	}

	// register all module
	intercept.middleware = make(types.MiddlewareGroup)
	intercept.opt = &server.opt
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
