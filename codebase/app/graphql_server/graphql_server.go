package graphqlserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"reflect"

	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/candishared"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/config"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/tracer"

	graphql "github.com/golangid/graphql-go"
	"github.com/graph-gophers/graphql-transport-ws/graphqlws"
)

const (
	rootGraphQLPath       = "/graphql"
	rootGraphQLPlayground = "/graphql/playground"
	rootGraphQLVoyager    = "/graphql/voyager"
)

type graphqlServer struct {
	httpEngine *http.Server
}

// NewServer create new GraphQL server
func NewServer(service factory.ServiceFactory) factory.AppServerFactory {

	httpEngine := new(http.Server)
	httpHandler := NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc(rootGraphQLPath, httpHandler.ServeGraphQL())
	mux.HandleFunc(rootGraphQLPlayground, httpHandler.ServePlayground)
	mux.HandleFunc(rootGraphQLVoyager, httpHandler.ServeVoyager)

	httpEngine.Addr = fmt.Sprintf(":%d", config.BaseEnv().HTTPPort)
	httpEngine.Handler = mux

	logger.LogYellow("[GraphQL] endpoint : " + rootGraphQLPath)
	logger.LogYellow("[GraphQL] playground : " + rootGraphQLPlayground)
	logger.LogYellow("[GraphQL] voyager : " + rootGraphQLVoyager)
	fmt.Printf("\x1b[34;1m⇨ GraphQL HTTP server run at port [::]%s\x1b[0m\n\n", httpEngine.Addr)

	return &graphqlServer{
		httpEngine: httpEngine,
	}
}

func (s *graphqlServer) Serve() {
	if err := s.httpEngine.ListenAndServe(); err != nil {
		switch e := err.(type) {
		case *net.OpError:
			panic(fmt.Errorf("gql http server: %v", e))
		}
	}
}

func (s *graphqlServer) Shutdown(ctx context.Context) {
	log.Println("Stopping GraphQL HTTP server...")
	defer func() { log.Println("Stopping GraphQL HTTP server: \x1b[32;1mSUCCESS\x1b[0m") }()

	s.httpEngine.Shutdown(ctx)
}

// Handler interface
type Handler interface {
	ServeGraphQL() http.HandlerFunc
	ServePlayground(resp http.ResponseWriter, req *http.Request)
	ServeVoyager(resp http.ResponseWriter, req *http.Request)
}

// NewHandler for create public graphql handler (maybe inject to rest handler)
func NewHandler(service factory.ServiceFactory) Handler {

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

	root.rootQuery = constructStruct(queryResolverFields, queryResolverValues)
	root.rootMutation = constructStruct(mutationResolverFields, mutationResolverValues)
	root.rootSubscription = constructStruct(subscriptionResolverFields, subscriptionResolverValues)
	gqlSchema := candihelper.LoadAllFile(config.BaseEnv().GraphQLSchemaDir, ".graphql")

	schemaOpts := []graphql.SchemaOpt{
		graphql.UseStringDescriptions(),
		graphql.UseFieldResolvers(),
		graphql.Tracer(&tracer.GraphQLTracer{}),
	}
	if config.BaseEnv().IsProduction {
		// handling vulnerabilities exploit schema
		schemaOpts = append(schemaOpts, graphql.DisableIntrospection())
	}
	schema := graphql.MustParseSchema(string(gqlSchema), &root, schemaOpts...)

	return &handlerImpl{
		schema: schema,
	}
}

type handlerImpl struct {
	schema *graphql.Schema
}

func (s *handlerImpl) ServeGraphQL() http.HandlerFunc {

	graphQLHandler := func(resp http.ResponseWriter, req *http.Request) {

		var params struct {
			Query         string                 `json:"query"`
			OperationName string                 `json:"operationName"`
			Variables     map[string]interface{} `json:"variables"`
		}
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(body, &params); err != nil {
			params.Query = string(body)
		}

		ip := req.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = req.Header.Get("X-Real-IP")
			if ip == "" {
				ip, _, _ = net.SplitHostPort(req.RemoteAddr)
			}
		}
		req.Header.Set("X-Real-IP", ip)

		ctx := context.WithValue(req.Context(), candishared.ContextKey("headers"), req.Header)
		response := s.schema.Exec(ctx, params.Query, params.OperationName, params.Variables)
		responseJSON, err := json.Marshal(response)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		resp.Header().Set("Content-Type", "application/json")
		resp.Write(responseJSON)
	}

	return graphqlws.NewHandlerFunc(s.schema, http.HandlerFunc(graphQLHandler))
}

func (s *handlerImpl) ServePlayground(resp http.ResponseWriter, req *http.Request) {
	if config.BaseEnv().IsProduction {
		http.Error(resp, "Forbidden", http.StatusForbidden)
		return
	}
	resp.Write([]byte(playgroundAsset))
}

func (s *handlerImpl) ServeVoyager(resp http.ResponseWriter, req *http.Request) {
	if config.BaseEnv().IsProduction {
		http.Error(resp, "Forbidden", http.StatusForbidden)
		return
	}
	resp.Write([]byte(voyagerAsset))
}
