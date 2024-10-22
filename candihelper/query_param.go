package candihelper

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func extractTagName(structField reflect.StructField, tags []string) (key string) {
	for _, tag := range tags {
		key = strings.TrimSuffix(structField.Tag.Get(tag), ",omitempty")
		if key == "-" {
			continue
		} else if key != "" {
			break
		}
	}
	if key == "" {
		key = structField.Name
	}
	return
}

// ParseFromQueryParam parse url query string to struct target (with multiple data type in struct field), target must in pointer
func ParseFromQueryParam(query URLQueryGetter, target any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	var parseDataTypeValue func(queryValue string, typ reflect.Type, val reflect.Value)

	var errs = NewMultiError()

	pValue := reflect.ValueOf(target)
	if pValue.Kind() != reflect.Ptr {
		panic(fmt.Errorf("%v is not pointer", pValue.Kind()))
	}
	pValue = pValue.Elem()
	pType := reflect.TypeOf(target).Elem()
	for i := 0; i < pValue.NumField(); i++ {
		field := pValue.Field(i)
		typ := pType.Field(i)
		if typ.Anonymous { // embedded struct
			if e, ok := ParseFromQueryParam(query, field.Addr().Interface()).(MultiError); ok {
				errs.Merge(e)
			}
		}

		key := extractTagName(typ, []string{"query", "json"})
		var v string
		if val := query.Get(key); val != "" {
			v = val
		} else {
			v = typ.Tag.Get("default")
		}

		parseDataTypeValue = func(queryValue string, sourceType reflect.Type, targetField reflect.Value) {
			switch sourceType.Kind() {
			case reflect.String:
				if ok, _ := strconv.ParseBool(typ.Tag.Get("lower")); ok {
					queryValue = strings.ToLower(queryValue)
				}
				targetField.SetString(queryValue)
			case reflect.Int32, reflect.Int, reflect.Int64:
				vInt, err := strconv.Atoi(queryValue)
				if queryValue != "" && err != nil {
					errs.Append(key, fmt.Errorf("Cannot parse '%s' (%T) to type number", queryValue, queryValue))
				}
				targetField.SetInt(int64(vInt))
			case reflect.Bool:
				vBool, err := strconv.ParseBool(queryValue)
				if queryValue != "" && err != nil {
					errs.Append(key, fmt.Errorf("Cannot parse '%s' (%T) to type boolean", queryValue, queryValue))
				}
				targetField.SetBool(vBool)
			case reflect.Float32, reflect.Float64:
				vFloat, err := strconv.ParseFloat(queryValue, 64)
				if queryValue != "" && err != nil {
					errs.Append(key, fmt.Errorf("Cannot parse '%s' (%T) to type float", queryValue, queryValue))
				}
				targetField.SetFloat(vFloat)
			case reflect.Slice:
				separator := typ.Tag.Get("separator")
				if separator == "" {
					separator = ","
				}
				values := strings.Split(queryValue, separator)
				targetSlice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(targetField.Interface()).Elem()), len(values), len(values))
				for i := 0; i < targetSlice.Len(); i++ {
					parseDataTypeValue(values[i], targetSlice.Index(i).Type(), targetSlice.Index(i))
				}
				targetField.Set(targetSlice)

			case reflect.Ptr:
				// allocate new value to pointer targetField
				targetField.Set(reflect.ValueOf(reflect.New(sourceType.Elem()).Interface()))
				parseDataTypeValue(queryValue, sourceType.Elem(), targetField.Elem())
			}
		}

		if v != "" {
			parseDataTypeValue(v, field.Type(), field)
		}
	}

	if errs.HasError() {
		return errs
	}

	return
}

// ParseToQueryParam parse struct data to query param
func ParseToQueryParam(source any) (s string) {
	defer func() { recover() }()

	pValue := ReflectValueUnwrapPtr(reflect.ValueOf(source))
	pType := ReflectTypeUnwrapPtr(reflect.TypeOf(source))

	var uri []string
	for i := 0; i < pValue.NumField(); i++ {
		field := ReflectValueUnwrapPtr(pValue.Field(i))
		typ := pType.Field(i)

		if typ.PkgPath != "" && !typ.Anonymous { // unexported
			continue
		}

		if typ.Anonymous { // embedded struct
			uri = append(uri, ParseToQueryParam(field.Interface()))
			continue
		}

		isOmitempty := strings.HasSuffix(typ.Tag.Get("json"), ",omitempty")
		key := extractTagName(typ, []string{"query", "json"})
		if key == "-" {
			continue
		}

		dataType := reflect.ValueOf(field.Interface()).Type().Kind()
		switch dataType {
		case reflect.String:
			val := url.PathEscape(field.String())
			if val == "" && isOmitempty {
				continue
			}
			uri = append(uri, fmt.Sprintf("%s=%s", key, val))
		case reflect.Ptr:
			val := ""
			if !field.IsNil() {
				val = fmt.Sprintf("%v", field.Elem())
			}
			uri = append(uri, fmt.Sprintf("%s=%s", key, url.PathEscape(val)))
		default:
			uri = append(uri, fmt.Sprintf("%s=%v", key, field.Interface()))
		}
	}

	return strings.Join(uri, "&")
}
