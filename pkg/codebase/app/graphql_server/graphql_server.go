package graphqlserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"

	graphql "github.com/golangid/graphql-go"
	"github.com/graph-gophers/graphql-transport-ws/graphqlws"
)

type graphqlServer struct {
	httpEngine  *http.Server
	httpHandler Handler
	service     factory.ServiceFactory
}

// NewServer create new GraphQL server
func NewServer(service factory.ServiceFactory) factory.AppServerFactory {
	return &graphqlServer{
		httpHandler: NewHandler(service),
	}
}

func (s *graphqlServer) Serve() {
	s.httpEngine = new(http.Server)

	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", s.httpHandler.ServeGraphQL())
	mux.HandleFunc("/graphql/playground", s.httpHandler.ServePlayground)
	mux.HandleFunc("/graphql/voyager", s.httpHandler.ServeVoyager)

	s.httpEngine.Addr = fmt.Sprintf(":%d", config.BaseEnv().HTTPPort)
	s.httpEngine.Handler = mux

	logger.LogYellow("[GraphQL] endpoint : /graphql")
	logger.LogYellow("[GraphQL] playground : /graphql/playground")
	logger.LogYellow("[GraphQL] voyager : /graphql/voyager")
	fmt.Printf("\x1b[34;1m⇨ GraphQL server run at port [::]%s\x1b[0m\n\n", s.httpEngine.Addr)
	if err := s.httpEngine.ListenAndServe(); err != nil {
		switch e := err.(type) {
		case *net.OpError:
			panic(fmt.Errorf("gql http server: %v", e))
		}
	}
}

func (s *graphqlServer) Shutdown(ctx context.Context) {
	deferFunc := logger.LogWithDefer("Stopping GraphQL HTTP server...")
	defer deferFunc()

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
	gqlSchema := helper.LoadAllFile(config.BaseEnv().GraphQLSchemaDir, ".graphql")

	schema := graphql.MustParseSchema(string(gqlSchema), &root,
		graphql.UseStringDescriptions(),
		graphql.UseFieldResolvers(),
		graphql.Tracer(&tracer.GraphQLTracer{}))
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

	return graphqlws.NewHandlerFunc(s.schema, http.HandlerFunc(graphQLHandler))
}

func (s *handlerImpl) ServePlayground(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte(`<!DOCTYPE html>
	<html>
	<head>
		<meta charset=utf-8/>
		<meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
		<link rel="shortcut icon" href="https://graphcool-playground.netlify.com/favicon.png">
		<link rel="stylesheet" href="//cdn.jsdelivr.net/npm/graphql-playground-react@1.7.8/build/static/css/index.css"/>
		<link rel="shortcut icon" href="//cdn.jsdelivr.net/npm/graphql-playground-react@1.7.8/build/favicon.png"/>
		<script src="//cdn.jsdelivr.net/npm/graphql-playground-react@1.7.8/build/static/js/middleware.js"></script>
		<title>Playground</title>
	</head>
	<body>
	<style type="text/css">
		html { font-family: "Open Sans", sans-serif; overflow: hidden; }
		body { margin: 0; background: #172a3a; }
	</style>
	<div id="root"/>
	<script type="text/javascript">
		window.addEventListener('load', function (event) {
			const root = document.getElementById('root');
			root.classList.add('playgroundIn');
			const wsProto = location.protocol == 'https:' ? 'wss:' : 'ws:'
			GraphQLPlayground.init(root, {
				endpoint: location.protocol + '//' + location.host + '/graphql',
				subscriptionsEndpoint: wsProto + '//' + location.host + '/graphql',
				settings: {
					'request.credentials': 'same-origin'
				}
			})
		})
	</script>
	</body>
	</html>`))
}

func (s *handlerImpl) ServeVoyager(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte(`<!DOCTYPE html>
	<html>
	  <head>
		<style>
		  body {
			height: 100%;
			margin: 0;
			width: 100%;
			overflow: hidden;
		  }
		  #voyager {
			height: 100vh;
		  }
		</style>
	
		<!--
		  This GraphQL Voyager example depends on Promise and fetch, which are available in
		  modern browsers, but can be "polyfilled" for older browsers.
		  GraphQL Voyager itself depends on React DOM.
		  If you do not want to rely on a CDN, you can host these files locally or
		  include them directly in your favored resource bunder.
		-->
		<script src="https://cdn.jsdelivr.net/es6-promise/4.0.5/es6-promise.auto.min.js"></script>
		<script src="https://cdn.jsdelivr.net/fetch/0.9.0/fetch.min.js"></script>
		<script src="https://cdn.jsdelivr.net/npm/react@16/umd/react.production.min.js"></script>
		<script src="https://cdn.jsdelivr.net/npm/react-dom@16/umd/react-dom.production.min.js"></script>
	
		<!--
		  These two files are served from jsDelivr CDN, however you may wish to
		  copy them directly into your environment, or perhaps include them in your
		  favored resource bundler.
		 -->
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-voyager/dist/voyager.css" />
		<script src="https://cdn.jsdelivr.net/npm/graphql-voyager/dist/voyager.min.js"></script>
	  </head>
	  <body>
		<div id="voyager">Loading...</div>
		<script>
	
		  // Defines a GraphQL introspection fetcher using the fetch API. You're not required to
		  // use fetch, and could instead implement introspectionProvider however you like,
		  // as long as it returns a Promise
		  // Voyager passes introspectionQuery as an argument for this function
		  function introspectionProvider(introspectionQuery) {
			// This example expects a GraphQL server at the path /graphql.
			// Change this to point wherever you host your GraphQL server.
			return fetch(location.protocol + '//' + location.host + '/graphql', {
			  method: 'post',
			  headers: {
				'Accept': 'application/json',
				'Content-Type': 'application/json',
			  },
			  body: JSON.stringify({query: introspectionQuery}),
			  credentials: 'include',
			}).then(function (response) {
			  return response.text();
			}).then(function (responseBody) {
			  try {
				return JSON.parse(responseBody);
			  } catch (error) {
				return responseBody;
			  }
			});
		  }
	
		  // Render <Voyager /> into the body.
		  GraphQLVoyager.init(document.getElementById('voyager'), {
			introspection: introspectionProvider
		  });
		</script>
	  </body>
	</html>`))
}
