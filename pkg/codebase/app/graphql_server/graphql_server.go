package graphqlserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"

	"agungdwiprasetyo.com/backend-microservices/api"
	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"github.com/graph-gophers/graphql-go"
)

type graphqlServer struct {
	httpEngine  *http.Server
	httpHandler handler
	service     factory.ServiceFactory
}

// NewServer create new GraphQL server
func NewServer(service factory.ServiceFactory) factory.AppServerFactory {
	resolverModules := make(map[string]interface{})
	var resolverFields []reflect.StructField // for creating dynamic struct
	for _, m := range service.GetModules() {
		if name, handler := m.GraphQLHandler(); handler != nil {
			resolverModules[name] = handler
			resolverFields = append(resolverFields, reflect.StructField{
				Name: name,
				Type: reflect.TypeOf(handler),
			})
		}
	}

	resolverVal := reflect.New(reflect.StructOf(resolverFields)).Elem()
	for k, v := range resolverModules {
		val := resolverVal.FieldByName(k)
		val.Set(reflect.ValueOf(v))
	}

	resolver := resolverVal.Addr().Interface()
	gqlSchema := api.LoadGraphQLSchema(string(service.Name()))

	schema := graphql.MustParseSchema(gqlSchema, resolver,
		graphql.UseStringDescriptions(),
		graphql.UseFieldResolvers(),
		graphql.Logger(&logger.PanicLogger{}),
		graphql.Tracer(&logger.NoopTracer{}))

	return &graphqlServer{
		httpHandler: handler{schema: schema},
		httpEngine:  &http.Server{},
		service:     service,
	}
}

func (s *graphqlServer) Serve() {

	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", s.httpHandler.serveGraphQL)
	mux.HandleFunc("/graphql/playground", s.httpHandler.servePlayground)

	s.httpEngine.Addr = fmt.Sprintf(":%d", config.BaseEnv().GraphQLPort)
	s.httpEngine.Handler = mux

	fmt.Printf("\x1b[34;1m⇨ GraphQL server run at port [::]%s\x1b[0m\n\n", s.httpEngine.Addr)
	if err := s.httpEngine.ListenAndServe(); err != nil {
		switch e := err.(type) {
		case *net.OpError:
			panic(e)
		}
	}
}

func (s *graphqlServer) Shutdown(ctx context.Context) {
	deferFunc := logger.LogWithDefer("Stopping GraphQL HTTP server...")
	defer deferFunc()

	s.httpEngine.Shutdown(ctx)
}

type handler struct {
	schema *graphql.Schema
}

func (s *handler) serveGraphQL(resp http.ResponseWriter, req *http.Request) {

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

	ctx := context.WithValue(req.Context(), shared.ContextKey("headers"), req.Header)
	response := s.schema.Exec(ctx, params.Query, params.OperationName, params.Variables)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.Write(responseJSON)
}

func (s *handler) servePlayground(resp http.ResponseWriter, req *http.Request) {

	b, err := ioutil.ReadFile("web/graphql_playground.html")
	if err != nil {
		http.Error(resp, err.Error(), http.StatusNotFound)
		return
	}

	resp.Write(b)
}
