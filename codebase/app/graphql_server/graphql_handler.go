package graphqlserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/app/graphql_server/ws"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/graphql-go"
	gqltypes "github.com/golangid/graphql-go/types"
)

// Handler interface
type Handler interface {
	ServeGraphQL() http.HandlerFunc
	ServePlayground(resp http.ResponseWriter, req *http.Request)
	ServeVoyager(resp http.ResponseWriter, req *http.Request)
}

// ConstructHandlerFromService for create public graphql handler (maybe inject to rest handler)
func ConstructHandlerFromService(service factory.ServiceFactory, opt Option) Handler {
	gqlSchema := candihelper.LoadAllFile(os.Getenv(candihelper.WORKDIR)+"api/graphql", ".graphql")
	var resolver rootResolver

	if opt.rootResolver == nil {
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

				if schema := resolverModule.Schema(); schema != "" {
					gqlSchema = append(gqlSchema, schema+"\n"...)
				}
			}
		}
		resolver = rootResolver{
			rootQuery:        constructStruct(queryResolverFields, queryResolverValues),
			rootMutation:     constructStruct(mutationResolverFields, mutationResolverValues),
			rootSubscription: constructStruct(subscriptionResolverFields, subscriptionResolverValues),
		}
	} else {
		gqlSchema = append(gqlSchema, opt.rootResolver.Schema()+"\n"...)
		resolver = rootResolver{
			rootQuery:        opt.rootResolver.Query(),
			rootMutation:     opt.rootResolver.Mutation(),
			rootSubscription: opt.rootResolver.Subscription(),
		}
	}

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

	logger.LogYellow(fmt.Sprintf("[GraphQL] endpoint\t\t\t: http://127.0.0.1:%d%s", opt.httpPort, opt.RootPath))
	logger.LogYellow(fmt.Sprintf("[GraphQL] playground\t\t\t: http://127.0.0.1:%d%s/playground", opt.httpPort, opt.RootPath))
	logger.LogYellow(fmt.Sprintf("[GraphQL] playground (with explorer)\t: http://127.0.0.1:%d%s/playground?graphiql=true", opt.httpPort, opt.RootPath))
	logger.LogYellow(fmt.Sprintf("[GraphQL] voyager\t\t\t: http://127.0.0.1:%d%s/voyager", opt.httpPort, opt.RootPath))

	return &handlerImpl{
		schema: graphql.MustParseSchema((string(gqlSchema)), &resolver, schemaOpts...),
		option: opt,
	}
}

type handlerImpl struct {
	schema *graphql.Schema
	option Option
}

func NewHandler(schema *graphql.Schema, opt Option) Handler {
	return &handlerImpl{
		schema: schema,
		option: opt,
	}
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

		req.Header.Set(candihelper.HeaderXRealIP, extractRealIPHeader(req))

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
	if s.option.DisableIntrospection {
		http.Error(resp, "Forbidden", http.StatusForbidden)
		return
	}

	if ok, _ := strconv.ParseBool(req.URL.Query().Get("graphiql")); ok {
		resp.Write([]byte(`<!DOCTYPE html>
<html lang=en>
	<head>
		<meta charset=utf-8>
		<title>Candi GraphiQL</title>
		<link rel=icon href=https://raw.githubusercontent.com/dotansimha/graphql-yoga/main/website/public/favicon.ico>
		<link rel=stylesheet href=https://unpkg.com/@graphql-yoga/graphiql@3.0.10/dist/style.css>
	</head>
	<body id=body class=no-focus-outline>
		<noscript>You need to enable JavaScript to run this app.</noscript>
		<div id=root></div>
		<script type=module>
			import{renderYogaGraphiQL}from"https://storage.googleapis.com/agungdp/bin/candi/graphiql/yoga-graphiql.es.js";
			renderYogaGraphiQL(root, {
				endpoint: '` + s.option.RootPath + `',
				title: 'GraphiQL'
			});
		</script>
	</body>
</html>`))
		return
	}

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
			endpoint: location.protocol + '//' + location.host + '` + s.option.RootPath + `',
			subscriptionsEndpoint: wsProto + '//' + location.host + '` + s.option.RootPath + `',
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
	if s.option.DisableIntrospection {
		http.Error(resp, "Forbidden", http.StatusForbidden)
		return
	}
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
		return fetch(location.protocol + '//' + location.host + '` + s.option.RootPath + `', {
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

// panicLogger is the default logger used to log panics that occur during query execution
type panicLogger struct{}

// LogPanic is used to log recovered panic values that occur during query execution
// https://github.com/graph-gophers/graphql-go/blob/master/log/log.go#L19 + custom add log to trace
func (l *panicLogger) LogPanic(ctx context.Context, value interface{}) {
	const size = 2 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]

	tracer.Log(ctx, "gql_panic", value)
	tracer.Log(ctx, "gql_panic_trace", buf)
}

func extractRealIPHeader(req *http.Request) string {
	for _, header := range []string{candihelper.HeaderXForwardedFor, candihelper.HeaderXRealIP} {
		if ip := req.Header.Get(header); ip != "" {
			return ip
		}
	}

	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}
