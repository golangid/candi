package app

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"

	graphqlschema "agungdwiprasetyo.com/backend-microservices/api/graphql"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/graph-gophers/graphql-go"
	"github.com/labstack/echo"
)

// graphQLHandler graphql
func (a *App) graphqlHandler(mw middleware.Middleware) *graphqlHandler {
	resolverModules := make(map[string]interface{})
	var resolverFields []reflect.StructField
	for _, m := range a.modules {
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
	gqlSchema := graphqlschema.LoadSchema(string(a.serviceName))

	schema := graphql.MustParseSchema(gqlSchema, resolver,
		graphql.UseStringDescriptions(),
		graphql.UseFieldResolvers(),
		graphql.Logger(&shared.PanicLogger{}),
		graphql.Tracer(&shared.NoopTracer{}))

	return &graphqlHandler{
		schema: schema,
		mw:     mw,
	}
}

type graphqlHandler struct {
	schema *graphql.Schema
	mw     middleware.Middleware
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
	c.Response().Header().Set("WWW-Authenticate", `Basic realm=""`)
	if err := h.mw.BasicAuth(c.Request().Header.Get("Authorization")); err != nil {
		return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Unauthorized").JSON(c.Response())
	}

	b, err := ioutil.ReadFile("web/graphql_playground.html")
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusNotFound, err.Error()).JSON(c.Response())
	}

	_, err = c.Response().Write(b)
	return err
}
