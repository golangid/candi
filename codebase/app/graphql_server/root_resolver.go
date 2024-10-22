package graphqlserver

import "reflect"

// rootResolver root
type rootResolver struct {
	rootQuery        any
	rootMutation     any
	rootSubscription any
}

func (r *rootResolver) Query() any {
	return r.rootQuery
}
func (r *rootResolver) Mutation() any {
	return r.rootMutation
}
func (r *rootResolver) Subscription() any {
	return r.rootSubscription
}

// for creating dynamic struct utils

func appendStructField(fieldName string, fieldValue any, structFields *[]reflect.StructField) {
	if fieldValue == nil {
		return
	}
	*structFields = append(*structFields, reflect.StructField{
		Name: fieldName,
		Type: reflect.TypeOf(fieldValue),
	})
}

func constructStruct(structFields []reflect.StructField, structValue map[string]any) any {
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
