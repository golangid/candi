package candihelper

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
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
	"unsafe"
)

// ParseFromQueryParam parse url query string to struct target (with multiple data type in struct field), target must in pointer
func ParseFromQueryParam(query URLQueryGetter, target interface{}) (err error) {
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

		var key string
		for _, tag := range []string{"query", "json"} {
			key = strings.TrimSuffix(typ.Tag.Get(tag), ",omitempty")
			if key == "-" {
				continue
			} else if key != "" {
				break
			}
		}

		if key == "" {
			key = typ.Name
		}

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

		isOmitempty := strings.HasSuffix(pType.Field(i).Tag.Get("json"), ",omitempty")
		key := strings.TrimSuffix(pType.Field(i).Tag.Get("json"), ",omitempty")
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

// ToFloat32Ptr helper
func ToFloat32Ptr(f float32) *float32 {
	return &f
}

// ToTimePtr helper
func ToTimePtr(t time.Time) *time.Time {
	return &t
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

// PtrToFloat32 helper
func PtrToFloat32(ptr *float32) (f float32) {
	if ptr != nil {
		f = *ptr
	}
	return
}

// PtrToTime helper
func PtrToTime(ptr *time.Time) (t time.Time) {
	if ptr != nil {
		return *ptr
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

// ToBytes convert all types to json bytes
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
		return "xxxxx"
	}
	pass, ok := u.User.Password()
	if pass == "" || !ok {
		return "xxxxx"
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

		case []string:
			separator := typ.Tag.Get("separator")
			if separator == "" {
				separator = ","
			}
			field.Set(reflect.ValueOf(strings.Split(val, separator)))

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

// PrintJSON for show data in pretty JSON with stack trace
func PrintJSON(data interface{}) {
	buff, _ := json.Marshal(data)
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, buff, "", "     ")
	fmt.Println(prettyJSON.String())
}

// GenerateHMAC generate random string
func GenerateHMAC(salt, str string) string {
	key := []byte(salt)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateSHA1 generate SHA1
func GenerateSHA1(input []byte) string {
	h := sha1.New()
	h.Write(input)
	return hex.EncodeToString(h.Sum(nil))
}

// ToCamelCase helper
func ToCamelCase(str string) string {
	str = strings.TrimSpace(str)
	if str == "" {
		return str
	}

	var n strings.Builder
	n.Grow(len(str))
	capNext := false
	for i, v := range str {
		vIsCap := v >= 'A' && v <= 'Z'
		vIsLow := v >= 'a' && v <= 'z'
		if capNext {
			if vIsLow {
				v += 'A'
				v -= 'a'
			}
		} else if i == 0 {
			if vIsCap {
				v += 'a'
				v -= 'A'
			}
		}
		if vIsCap || vIsLow {
			n.WriteByte(byte(v))
			capNext = false
		} else if vIsNum := v >= '0' && v <= '9'; vIsNum {
			n.WriteByte(byte(v))
			capNext = true
		} else {
			capNext = v == '_' || v == ' ' || v == '-' || v == '.'
		}
	}
	return n.String()
}

// ToDelimited helper, can use to snake_case or kebab-case
func ToDelimited(s string, delimiter uint8) string {
	s = strings.TrimSpace(s)
	n := strings.Builder{}
	screaming, ignore := false, ""
	n.Grow(len(s) + 2)
	for i, v := range []byte(s) {
		vIsCap := v >= 'A' && v <= 'Z'
		vIsLow := v >= 'a' && v <= 'z'
		if vIsLow && screaming {
			v += 'A'
			v -= 'a'
		} else if vIsCap && !screaming {
			v += 'a'
			v -= 'A'
		}

		if i+1 < len(s) {
			next := s[i+1]
			vIsNum := v >= '0' && v <= '9'
			nextIsCap := next >= 'A' && next <= 'Z'
			nextIsLow := next >= 'a' && next <= 'z'
			nextIsNum := next >= '0' && next <= '9'
			if (vIsCap && (nextIsLow || nextIsNum)) || (vIsLow && (nextIsCap || nextIsNum)) || (vIsNum && (nextIsCap || nextIsLow)) {
				prevIgnore := ignore != "" && i > 0 && strings.ContainsAny(string(s[i-1]), ignore)
				if !prevIgnore {
					if vIsCap && nextIsLow {
						if prevIsCap := i > 0 && s[i-1] >= 'A' && s[i-1] <= 'Z'; prevIsCap {
							n.WriteByte(delimiter)
						}
					}
					n.WriteByte(v)
					if vIsLow || vIsNum || nextIsNum {
						n.WriteByte(delimiter)
					}
					continue
				}
			}
		}

		if (v == ' ' || v == '_' || v == '-' || v == '.') && !strings.ContainsAny(string(v), ignore) {
			n.WriteByte(delimiter)
		} else {
			n.WriteByte(v)
		}
	}

	return n.String()
}

// GetRuntimeStackLine helper
func GetRuntimeStackLine() string {

	var name, file string
	var line int
	var pc [16]uintptr

	n := runtime.Callers(2, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(pc)
		name = fn.Name()
		if !strings.HasPrefix(name, "runtime.") {
			break
		}
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// ToString helper
func ToString(val interface{}) (str string) {
	switch s := val.(type) {
	case string:
		return s
	case bool:
		return strconv.FormatBool(s)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32)
	case int:
		return strconv.Itoa(s)
	case int64:
		return strconv.FormatInt(s, 10)
	case int32:
		return strconv.Itoa(int(s))
	case int16:
		return strconv.FormatInt(int64(s), 10)
	case int8:
		return strconv.FormatInt(int64(s), 10)
	case uint:
		return strconv.FormatInt(int64(s), 10)
	case uint64:
		return strconv.FormatInt(int64(s), 10)
	case uint32:
		return strconv.FormatInt(int64(s), 10)
	case uint16:
		return strconv.FormatInt(int64(s), 10)
	case uint8:
		return strconv.FormatInt(int64(s), 10)
	case []byte:
		return string(s)
	case nil:
		return ""
	case fmt.Stringer:
		return s.String()
	case error:
		return s.Error()
	default:
		return ""
	}
}

// ToInt helper
func ToInt(val interface{}) (i int) {
	switch s := val.(type) {
	case int:
		return s
	case int64:
		return int(s)
	case int32:
		return int(s)
	case int16:
		return int(s)
	case int8:
		return int(s)
	case uint:
		return int(s)
	case uint64:
		return int(s)
	case uint32:
		return int(s)
	case uint16:
		return int(s)
	case uint8:
		return int(s)
	case float64:
		return int(s)
	case float32:
		return int(s)
	case string:
		v, err := strconv.ParseInt(s, 0, 0)
		if err == nil {
			return int(v)
		}
		return 0
	case bool:
		if s {
			return 1
		}
		return 0
	case nil:
		return 0
	default:
		return 0
	}
}

// ByteToString helper with zero allocation
func ByteToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToByte helper with zero allocation
func StringToByte(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

// ParseTimeToString helper, return empty string if zero time
func ParseTimeToString(date time.Time, format string) (res string) {
	if !date.IsZero() {
		res = date.Format(format)
	}
	return res
}

// TransformSizeToByte helper
func TransformSizeToByte(size uint64) string {
	var unit string
	if size >= TByte {
		unit = "TB"
		size /= TByte
	} else if size >= GByte {
		unit = "GB"
		size /= GByte
	} else if size >= MByte {
		unit = "MB"
		size /= MByte
	} else if size >= KByte {
		unit = "KB"
		size /= KByte
	} else {
		unit = "B"
		size /= Byte
	}

	return fmt.Sprintf("%d %s", size, unit)
}

// UnwrapPtr take value from pointer
func UnwrapPtr[T any](t *T) (res T) {
	if t == nil {
		return
	}

	res = *t
	return
}

// WrapPtr set pointer from value
func WrapPtr[T any](t T) *T {
	return &t
}

// ToMap transform slice to map
func ToMap[T any, K comparable](list []T, keyGetter func(T) K) map[K]T {
	mp := make(map[K]T, len(list))
	for _, el := range list {
		mp[keyGetter(el)] = el
	}
	return mp
}

// IsExistInMap check key is exist in map
func IsExistInMap[T any, K comparable](m map[K]T, key K) bool {
	_, ok := m[key]
	return ok
}

// ToKeyMapSlice transform key of map to slice
func ToKeyMapSlice[T any, K comparable](mp map[K]T) (list []K) {
	for k := range mp {
		list = append(list, k)
	}
	return list
}
