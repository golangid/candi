package app

import (
	"fmt"
	"log"
	"net"

	"github.com/agungdwiprasetyo/backend-microservices/config"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/helper"
)

// ServeGRPC user service
func (a *App) ServeGRPC() {
	if !config.GlobalEnv.UseGRPC {
		return
	}

	grpcPort := fmt.Sprintf(":%d", config.GlobalEnv.GRPCPort)

	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s⇨ Server Run at port [::]%s\n\n", helper.GRPCBanner, grpcPort)

	// register all module
	for _, m := range a.modules {
		if h := m.GRPCHandler(); h != nil {
			h.Register(a.grpcServer)
		}
	}

	err = a.grpcServer.Serve(listener)
	if err != nil {
		log.Println("Unexpected Error", err)
	}
}
