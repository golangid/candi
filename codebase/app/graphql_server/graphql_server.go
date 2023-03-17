package graphqlserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/app/graphql_server/static"
	"github.com/golangid/candi/codebase/app/graphql_server/ws"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"

	graphql "github.com/golangid/graphql-go"
	gqltypes "github.com/golangid/graphql-go/types"
	"github.com/soheilhy/cmux"
)

const (
	rootGraphQLPath       = "/graphql"
	rootGraphQLPlayground = "/graphql/playground"
	rootGraphQLVoyager    = "/graphql/voyager"
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

	httpHandler := NewHandler(service, server.opt)

	mux := http.NewServeMux()
	mux.Handle("/", server.opt.rootHandler)
	mux.Handle("/memstats", service.GetDependency().GetMiddleware().HTTPBasicAuth(http.HandlerFunc(wrapper.HTTPHandlerMemstats)))
	mux.HandleFunc(server.opt.rootPath+rootGraphQLPath, httpHandler.ServeGraphQL())
	mux.HandleFunc(server.opt.rootPath+rootGraphQLPlayground, httpHandler.ServePlayground)
	mux.HandleFunc(server.opt.rootPath+rootGraphQLVoyager, httpHandler.ServeVoyager)

	httpEngine.Addr = server.opt.httpPort
	httpEngine.Handler = mux

	logger.LogYellow("[GraphQL] endpoint : " + server.opt.rootPath + rootGraphQLPath)
	logger.LogYellow("[GraphQL] playground : " + server.opt.rootPath + rootGraphQLPlayground)
	logger.LogYellow("[GraphQL] voyager : " + server.opt.rootPath + rootGraphQLVoyager)
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

	switch e := err.(type) {
	case *net.OpError:
		panic(fmt.Errorf("gql http server: %v", e))
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

// Handler interface
type Handler interface {
	ServeGraphQL() http.HandlerFunc
	ServePlayground(resp http.ResponseWriter, req *http.Request)
	ServeVoyager(resp http.ResponseWriter, req *http.Request)
}

// NewHandler for create public graphql handler (maybe inject to rest handler)
func NewHandler(service factory.ServiceFactory, opt Option) Handler {

	// create dynamic struct
	queryResolverValues := make(map[string]interface{})
	mutationResolverValues := make(map[string]interface{})
	subscriptionResolverValues := make(map[string]interface{})
	var queryResolverFields, mutationResolverFields, subscriptionResolverFields []reflect.StructField
	for _, m := range service.GetModules() {
		if resolverModule := m.GraphQLHandler(); resolverModule != nil {
			rootName := string(m.Name())
			query, mutation, subscription := resolverModule.Query(), resolverModule.Mutation(), resolverModule.Subscription()

			appendStructField(rootName, query, &queryResolverFields)
			appendStructField(rootName, mutation, &mutationResolverFields)
			appendStructField(rootName, subscription, &subscriptionResolverFields)

			queryResolverValues[rootName] = query
			mutationResolverValues[rootName] = mutation
			subscriptionResolverValues[rootName] = subscription
		}
	}

	opt.rootResolver.rootQuery = constructStruct(queryResolverFields, queryResolverValues)
	opt.rootResolver.rootMutation = constructStruct(mutationResolverFields, mutationResolverValues)
	opt.rootResolver.rootSubscription = constructStruct(subscriptionResolverFields, subscriptionResolverValues)
	gqlSchema := candihelper.LoadAllFile(os.Getenv(candihelper.WORKDIR)+"api/graphql", ".graphql")

	// default directive
	directiveFuncs := map[string]gqltypes.DirectiveFunc{
		"auth":          service.GetDependency().GetMiddleware().GraphQLAuth,
		"permissionACL": service.GetDependency().GetMiddleware().GraphQLPermissionACL,
	}
	for directive, dirFunc := range opt.directiveFuncs {
		directiveFuncs[directive] = dirFunc
	}

	schemaOpts := []graphql.SchemaOpt{
		graphql.UseStringDescriptions(),
		graphql.UseFieldResolvers(),
		graphql.Tracer(&graphqlTracer{}),
		graphql.Logger(&panicLogger{}),
		graphql.DirectiveFuncs(directiveFuncs),
	}
	if opt.DisableIntrospection {
		// handling vulnerabilities exploit schema
		schemaOpts = append(schemaOpts, graphql.DisableIntrospection())
	}
	schema := graphql.MustParseSchema(string(gqlSchema), &opt.rootResolver, schemaOpts...)

	return &handlerImpl{
		disableIntrospection: opt.DisableIntrospection,
		schema:               schema,
	}
}

type handlerImpl struct {
	disableIntrospection bool
	schema               *graphql.Schema
}

func (s *handlerImpl) ServeGraphQL() http.HandlerFunc {

	return ws.NewHandlerFunc(s.schema, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {

		var params struct {
			Query         string                 `json:"query"`
			OperationName string                 `json:"operationName"`
			Variables     map[string]interface{} `json:"variables"`
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(body, &params); err != nil {
			params.Query = string(body)
		}

		ip := req.Header.Get(candihelper.HeaderXForwardedFor)
		if ip == "" {
			ip = req.Header.Get(candihelper.HeaderXRealIP)
			if ip == "" {
				ip, _, _ = net.SplitHostPort(req.RemoteAddr)
			}
		}
		req.Header.Set(candihelper.HeaderXRealIP, ip)

		ctx := context.WithValue(req.Context(), candishared.ContextKeyHTTPHeader, req.Header)
		response := s.schema.Exec(ctx, params.Query, params.OperationName, params.Variables)
		responseJSON, err := json.Marshal(response)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		resp.Header().Set(candihelper.HeaderContentType, candihelper.HeaderMIMEApplicationJSON)
		resp.Write(responseJSON)
	}))
}

func (s *handlerImpl) ServePlayground(resp http.ResponseWriter, req *http.Request) {
	if s.disableIntrospection {
		http.Error(resp, "Forbidden", http.StatusForbidden)
		return
	}
	resp.Write([]byte(static.PlaygroundAsset))
}

func (s *handlerImpl) ServeVoyager(resp http.ResponseWriter, req *http.Request) {
	if s.disableIntrospection {
		http.Error(resp, "Forbidden", http.StatusForbidden)
		return
	}
	resp.Write([]byte(static.VoyagerAsset))
}

// panicLogger is the default logger used to log panics that occur during query execution
type panicLogger struct{}

// LogPanic is used to log recovered panic values that occur during query execution
// https://github.com/graph-gophers/graphql-go/blob/master/log/log.go#L19 + custom add log to trace
func (l *panicLogger) LogPanic(ctx context.Context, value interface{}) {
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]

	tracer.Log(ctx, "gql_panic", value)
	tracer.Log(ctx, "gql_panic_trace", buf)
}
