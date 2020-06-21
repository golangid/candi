package app

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"

	"agungdwiprasetyo.com/backend-microservices/api"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/graph-gophers/graphql-go"
	"github.com/labstack/echo"
)

// graphQLHandler graphql
func (a *App) graphqlHandler() *graphqlHandler {
	resolverModules := make(map[string]interface{})
	var resolverFields []reflect.StructField // for creating dynamic struct
	for _, m := range a.service.GetModules() {
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
	gqlSchema := api.LoadGraphQLSchema(string(a.service.Name()))

	schema := graphql.MustParseSchema(gqlSchema, resolver,
		graphql.UseStringDescriptions(),
		graphql.UseFieldResolvers(),
		graphql.Logger(&logger.PanicLogger{}),
		graphql.Tracer(&logger.NoopTracer{}))

	return &graphqlHandler{
		schema: schema,
	}
}

type graphqlHandler struct {
	schema *graphql.Schema
	mw     interfaces.Middleware
}

func (h *graphqlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &params); err != nil {
		params.Query = string(body)
	}

	ip := r.Header.Get(echo.HeaderXForwardedFor)
	if ip == "" {
		ip = r.Header.Get(echo.HeaderXRealIP)
		if ip == "" {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
		}
	}
	r.Header.Set(echo.HeaderXRealIP, ip)

	ctx := context.WithValue(r.Context(), shared.ContextKey("headers"), r.Header)
	response := h.schema.Exec(ctx, params.Query, params.OperationName, params.Variables)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func (h *graphqlHandler) servePlayground(c echo.Context) error {

	b, err := ioutil.ReadFile("web/graphql_playground.html")
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusNotFound, err.Error()).JSON(c.Response())
	}

	_, err = c.Response().Write(b)
	return err
}
