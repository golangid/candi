package graphqlserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/wrapper"

	"github.com/soheilhy/cmux"
)

type graphqlServer struct {
	opt        Option
	httpEngine *http.Server
	listener   net.Listener
}

// NewServer create new GraphQL server
func NewServer(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {

	httpEngine := new(http.Server)
	server := &graphqlServer{
		httpEngine: httpEngine,
		opt:        getDefaultOption(),
	}
	for _, opt := range opts {
		opt(&server.opt)
	}

	httpHandler := ConstructHandlerFromService(service, server.opt)

	mux := http.NewServeMux()
	mux.Handle("/", server.opt.rootHandler)
	mux.Handle("/memstats", service.GetDependency().GetMiddleware().HTTPBasicAuth(http.HandlerFunc(wrapper.HTTPHandlerMemstats)))
	mux.HandleFunc(server.opt.RootPath, httpHandler.ServeGraphQL())
	mux.HandleFunc(server.opt.RootPath+"/playground", httpHandler.ServePlayground)
	mux.HandleFunc(server.opt.RootPath+"/voyager", httpHandler.ServeVoyager)

	httpEngine.Addr = fmt.Sprintf(":%d", server.opt.httpPort)
	httpEngine.Handler = mux

	fmt.Printf("\x1b[34;1m⇨ GraphQL HTTP server run at port [::]%s\x1b[0m\n\n", httpEngine.Addr)

	if server.opt.sharedListener != nil {
		server.listener = server.opt.sharedListener.Match(cmux.HTTP1Fast(http.MethodPatch))
	}

	return server
}

func (s *graphqlServer) Serve() {
	var err error
	if s.listener != nil {
		err = s.httpEngine.Serve(s.listener)
	} else {
		err = s.httpEngine.ListenAndServe()
	}

	switch err.(type) {
	case *net.OpError:
		log.Panicf("GraphQL Server: Unexpected Error: %v", err)
	}
}

func (s *graphqlServer) Shutdown(ctx context.Context) {
	defer log.Println("\x1b[33;1mStopping GraphQL HTTP server:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")

	s.httpEngine.Shutdown(ctx)
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *graphqlServer) Name() string {
	return string(types.GraphQL)
}
