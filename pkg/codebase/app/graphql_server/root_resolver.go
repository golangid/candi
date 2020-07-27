package graphqlserver

import "reflect"

// issue https://github.com/graph-gophers/graphql-go/issues/145
type rootResolver struct {
	rootQuery        interface{}
	rootMutation     interface{}
	rootSubscription interface{}
}

func (r *rootResolver) Query() interface{} {
	return r.rootQuery
}
func (r *rootResolver) Mutation() interface{} {
	return r.rootMutation
}
func (r *rootResolver) Subscription() interface{} {
	return r.rootSubscription
}

var root rootResolver

// SetRootSubscription public function
// this public method made because cannot create dynamic method for embedded struct (issue https://github.com/golang/go/issues/15924)
// and subscription in graphql cannot subscribe to at most one subscription at a time
func SetRootSubscription(subsRootResolver interface{}) {
	root.rootSubscription = subsRootResolver
}

// for creating dynamic struct utils

func appendStructField(fieldName string, fieldValue interface{}, structFields *[]reflect.StructField) {
	if fieldValue == nil {
		return
	}
	*structFields = append(*structFields, reflect.StructField{
		Name: fieldName,
		Type: reflect.TypeOf(fieldValue),
	})
}

func constructStruct(structFields []reflect.StructField, structValue map[string]interface{}) interface{} {
	if len(structFields) == 0 {
		return &struct{}{}
	}

	structType := reflect.New(reflect.StructOf(structFields)).Elem()
	for fieldName, fieldValue := range structValue {
		val := structType.FieldByName(fieldName)
		val.Set(reflect.ValueOf(fieldValue))
	}
	return structType.Addr().Interface()
}
