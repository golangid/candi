package graphqlserver

import "reflect"

// rootResolver root
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
		if fieldValue == nil {
			continue
		}
		val := structType.FieldByName(fieldName)
		val.Set(reflect.ValueOf(fieldValue))
	}
	return structType.Addr().Interface()
}
