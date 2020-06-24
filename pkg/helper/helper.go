package helper

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// ParseFromQueryParam parse url query string to struct target (with multiple data type in struct field), target must in pointer
func ParseFromQueryParam(query url.Values, target interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	var parseDataTypeValue func(typ reflect.Type, val reflect.Value)

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
			if e, ok := ParseFromQueryParam(query, field.Addr().Interface()).(*MultiError); ok {
				errs.Merge(e)
			}
		}

		key := strings.TrimSuffix(typ.Tag.Get("json"), ",omitempty")
		if key == "-" {
			continue
		}

		var v string
		if val := query[key]; len(val) > 0 && val[0] != "" {
			v = val[0]
		} else {
			v = typ.Tag.Get("default")
		}

		parseDataTypeValue = func(sourceType reflect.Type, targetField reflect.Value) {
			switch sourceType.Kind() {
			case reflect.String:
				if ok, _ := strconv.ParseBool(typ.Tag.Get("lower")); ok {
					v = strings.ToLower(v)
				}
				targetField.SetString(v)
			case reflect.Int32, reflect.Int, reflect.Int64:
				vInt, err := strconv.Atoi(v)
				if v != "" && err != nil {
					errs.Append(key, fmt.Errorf("Cannot parse '%s' (%T) to type number", v, v))
				}
				targetField.SetInt(int64(vInt))
			case reflect.Bool:
				vBool, err := strconv.ParseBool(v)
				if v != "" && err != nil {
					errs.Append(key, fmt.Errorf("Cannot parse '%s' (%T) to type boolean", v, v))
				}
				targetField.SetBool(vBool)
			case reflect.Ptr:
				if v != "" {
					// allocate new value to pointer targetField
					targetField.Set(reflect.ValueOf(reflect.New(sourceType.Elem()).Interface()))
					parseDataTypeValue(sourceType.Elem(), targetField.Elem())
				}
			}
		}

		parseDataTypeValue(field.Type(), field)
	}

	if errs.HasError() {
		return errs
	}

	return
}

// StringYellow func
func StringYellow(str string) string {
	return fmt.Sprintf("\x1b[33;2m%s\x1b[0m", str)
}

// StringGreen func
func StringGreen(str string) string {
	return fmt.Sprintf("\x1b[32;2m%s\x1b[0m", str)
}

// ToBoolPtr helper
func ToBoolPtr(b bool) *bool {
	return &b
}

// ToStringPtr helper
func ToStringPtr(str string) *string {
	return &str
}

// CronJobKeyToString helper
func CronJobKeyToString(jobName string, interval string) string {
	return fmt.Sprintf("%s:%s", jobName, interval)
}

// ParseCronJobKey helper
func ParseCronJobKey(str string) (jobName string, interval string) {
	split := strings.Split(str, ":")
	jobName = split[0]
	interval = split[1]
	return
}
