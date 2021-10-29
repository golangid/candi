package candihelper

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
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
			if e, ok := ParseFromQueryParam(query, field.Addr().Interface()).(MultiError); ok {
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

// ParseToQueryParam parse struct data to query param
func ParseToQueryParam(source interface{}) (s string) {
	defer func() { recover() }()

	pValue := reflect.ValueOf(source)
	pType := reflect.TypeOf(source)
	if pValue.Kind() == reflect.Ptr {
		pValue = pValue.Elem()
		pType = pType.Elem()
	}

	var uri []string
	for i := 0; i < pValue.NumField(); i++ {
		field := pValue.Field(i)
		typ := pType.Field(i)

		if typ.PkgPath != "" && !typ.Anonymous { // unexported
			continue
		}

		if typ.Anonymous { // embedded struct
			uri = append(uri, ParseToQueryParam(field.Interface()))
			continue
		}
		key := strings.TrimSuffix(pType.Field(i).Tag.Get("json"), ",omitempty")
		if key == "-" {
			continue
		}

		switch field.Interface().(type) {
		case string:
			val := url.PathEscape(field.String())
			uri = append(uri, fmt.Sprintf("%s=%s", key, val))
		default:
			uri = append(uri, fmt.Sprintf("%s=%v", key, field.Interface()))
		}
	}

	return strings.Join(uri, "&")
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

// ToIntPtr helper
func ToIntPtr(i int) *int {
	return &i
}

// ToFloatPtr helper
func ToFloatPtr(f float64) *float64 {
	return &f
}

// PtrToString helper
func PtrToString(ptr *string) (s string) {
	if ptr != nil {
		s = *ptr
	}
	return
}

// PtrToBool helper
func PtrToBool(ptr *bool) (b bool) {
	if ptr != nil {
		b = *ptr
	}
	return
}

// PtrToInt helper
func PtrToInt(ptr *int) (i int) {
	if ptr != nil {
		i = *ptr
	}
	return
}

// PtrToFloat helper
func PtrToFloat(ptr *float64) (f float64) {
	if ptr != nil {
		f = *ptr
	}
	return
}

// ToAsiaJakartaTime convert only time location to AsiaJakarta local time
func ToAsiaJakartaTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(),
		t.Nanosecond(), AsiaJakartaLocalTime)
}

// ToUTC convert only location to UTC
func ToUTC(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(),
		t.Nanosecond(), time.UTC)
}

// TimeRemoveNanosecond ...
func TimeRemoveNanosecond(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(),
		0, t.Location())
}

// ToBytes convert all types to bytes
func ToBytes(i interface{}) (b []byte) {
	switch t := i.(type) {
	case []byte:
		b = t
	case string:
		b = []byte(t)
	default:
		b, _ = json.Marshal(i)
	}
	return
}

// StringInSlice function for checking whether string in slice
// str string searched string
// list []string slice
func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// MaskingPasswordURL for hide plain text password from given URL format
func MaskingPasswordURL(stringURL string) string {
	u, err := url.Parse(stringURL)
	if err != nil {
		return stringURL
	}
	pass, ok := u.User.Password()
	if pass == "" || !ok {
		return stringURL
	}

	u.User = url.UserPassword(u.User.Username(), "xxxxx")
	return u.String()
}

// MustParseEnv must parse env to struct, panic if env from target struct tag is not found
func MustParseEnv(target interface{}) {
	pValue := reflect.ValueOf(target)
	pValue = pValue.Elem()
	pType := reflect.TypeOf(target).Elem()
	mErrs := NewMultiError()
	for i := 0; i < pValue.NumField(); i++ {
		field := pValue.Field(i)
		if !field.CanSet() { // skip if field cannot set a value (usually an unexported field in struct), to avoid a panic
			continue
		}

		typ := pType.Field(i)
		if typ.Anonymous ||
			(typ.Type.Kind() == reflect.Struct && !reflect.DeepEqual(field.Interface(), time.Time{})) { // embedded struct or struct field
			MustParseEnv(field.Addr().Interface())
			continue
		}

		key := typ.Tag.Get("env")
		if key == "" || key == "-" {
			continue
		}

		val, ok := os.LookupEnv(key)
		if !ok {
			mErrs.Append(key, fmt.Errorf("missing %s environment", key))
			continue
		}

		switch field.Interface().(type) {
		case time.Duration:
			dur, err := time.ParseDuration(val)
			if err != nil {
				mErrs.Append(key, fmt.Errorf("env '%s': %v", key, err))
				continue
			}
			field.Set(reflect.ValueOf(dur))
		case time.Time:
			t, err := time.Parse(time.RFC3339, val)
			if err != nil {
				mErrs.Append(key, fmt.Errorf("env '%s': %v", key, err))
				continue
			}
			field.Set(reflect.ValueOf(t))
		case int32, int, int64:
			vInt, err := strconv.Atoi(val)
			if err != nil {
				mErrs.Append(key, fmt.Errorf("env '%s': %v", key, err))
				continue
			}
			field.SetInt(int64(vInt))
		case float32, float64:
			vFloat, err := strconv.ParseFloat(val, 64)
			if err != nil {
				mErrs.Append(key, fmt.Errorf("env '%s': %v", key, err))
				continue
			}
			field.SetFloat(vFloat)
		case bool:
			vBool, err := strconv.ParseBool(val)
			if err != nil {
				mErrs.Append(key, fmt.Errorf("env '%s': %v", key, err))
				continue
			}
			field.SetBool(vBool)
		case string:
			field.SetString(val)
		}
	}

	if mErrs.HasError() {
		panic("Environment error: \n" + mErrs.Error())
	}
}

// GetFuncName get function name in string
func GetFuncName(fn interface{}) string {
	defer func() { recover() }()
	handlerName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	return strings.TrimSuffix(strings.TrimPrefix(filepath.Ext(handlerName), "."), "-fm") // if `fn` is method, trim `-fm`
}
