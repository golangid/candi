package app

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	graphqlschema "agungdwiprasetyo.com/backend-microservices/api/graphql"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"github.com/graph-gophers/graphql-go"
)

// graphQLHandler graphql
func (a *App) graphqlHandler() *graphqlHandler {
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

	return &graphqlHandler{
		schema: graphql.MustParseSchema(gqlSchema, resolver,
			graphql.UseStringDescriptions(),
			graphql.UseFieldResolvers(),
			graphql.Logger(&shared.PanicLogger{}),
			graphql.Tracer(&shared.NoopTracer{})),
	}
}

type graphqlHandler struct {
	schema *graphql.Schema
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
