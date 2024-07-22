package candishared

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
)

type partialUpdateOption struct {
	updateFields map[string]struct{}
	ignoreFields map[string]struct{}
}

// DBUpdateOptionFunc option func
type DBUpdateOptionFunc func(*partialUpdateOption)

// DBUpdateOptionKeyExtractorResult option func
type DBUpdateOptionKeyExtractorResult struct {
	Key           string
	IsDefaultNull bool
	MustSet       bool
}

// DBUpdateSetUpdatedFields option func
func DBUpdateSetUpdatedFields(fields ...string) DBUpdateOptionFunc {
	return func(o *partialUpdateOption) {
		o.updateFields = make(map[string]struct{})
		for _, field := range fields {
			o.updateFields[field] = struct{}{}
		}
	}
}

// DBUpdateSetIgnoredFields option func
func DBUpdateSetIgnoredFields(fields ...string) DBUpdateOptionFunc {
	return func(o *partialUpdateOption) {
		o.ignoreFields = make(map[string]struct{})
		for _, field := range fields {
			o.ignoreFields[field] = struct{}{}
		}
	}
}

// DBUpdateGORMExtractorKey struct field key extractor for gorm model
func DBUpdateGORMExtractorKey(structField reflect.StructField) (res DBUpdateOptionKeyExtractorResult) {
	gormTag := structField.Tag.Get("gorm")
	res.IsDefaultNull = strings.Contains(strings.ToLower(gormTag), "default:null")
	res.Key = candihelper.ToDelimited(structField.Name, '_')
	if strings.HasPrefix(gormTag, "column:") {
		res.Key = strings.Split(strings.TrimPrefix(gormTag, "column:"), ";")[0]
	}
	return res
}

// DBUpdateMongoExtractorKey struct field key extractor for mongo model
func DBUpdateMongoExtractorKey(structField reflect.StructField) (res DBUpdateOptionKeyExtractorResult) {
	res.Key = candihelper.ToDelimited(structField.Name, '_')
	if bsonTag := strings.TrimSuffix(structField.Tag.Get("bson"), ",omitempty"); bsonTag != "" {
		res.Key = bsonTag
		res.MustSet = true
	}
	return res
}

// DBUpdateSqlExtractorKey struct field key extractor for mongo model
func DBUpdateSqlExtractorKey(structField reflect.StructField) (res DBUpdateOptionKeyExtractorResult) {
	sqlTag := structField.Tag.Get("sql")
	res.Key = candihelper.ToDelimited(structField.Name, '_')
	if strings.HasPrefix(sqlTag, "column:") {
		res.Key = strings.Split(strings.TrimPrefix(sqlTag, "column:"), ";")[0]
	}
	return res
}

// DBUpdateTools for construct selected field to update
type DBUpdateTools struct {
	KeyExtractorFunc    func(structTag reflect.StructField) (res DBUpdateOptionKeyExtractorResult)
	FieldValueExtractor func(reflect.Value) (val any, skip bool)
	IgnoredFields       []string
}

func (d *DBUpdateTools) parseOption(opts ...DBUpdateOptionFunc) (o partialUpdateOption) {
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// ToMap method
func (d DBUpdateTools) ToMap(data interface{}, opts ...DBUpdateOptionFunc) map[string]interface{} {
	opt := d.parseOption(opts...)

	dataValue := candihelper.ReflectValueUnwrapPtr(reflect.ValueOf(data))
	dataType := candihelper.ReflectTypeUnwrapPtr(reflect.TypeOf(data))
	isPartial := len(opt.updateFields) > 0 || len(opt.ignoreFields) > 0

	updateFields := make(map[string]interface{}, 0)
	for i := 0; i < dataValue.NumField(); i++ {
		fieldValue := candihelper.ReflectValueUnwrapPtr(dataValue.Field(i))
		fieldType := dataType.Field(i)

		if fieldType.Anonymous {
			for k, v := range d.ToMap(fieldValue.Interface(), opts...) {
				updateFields[k] = v
			}
			continue
		}

		var mustSet, isDefaultNull bool
		key := strings.TrimSuffix(fieldType.Tag.Get("json"), ",omitempty")
		if d.KeyExtractorFunc != nil {
			keyExtractor := d.KeyExtractorFunc(fieldType)
			key, isDefaultNull, mustSet = keyExtractor.Key, keyExtractor.IsDefaultNull, keyExtractor.MustSet
		}
		isIgnore, _ := strconv.ParseBool(fieldType.Tag.Get("ignoreUpdate"))
		if key == "" || key == "-" || isIgnore {
			continue
		}

		val, skipSet := fieldValue.Interface(), false

		if d.FieldValueExtractor != nil {
			val, skipSet = d.FieldValueExtractor(dataValue.Field(i))
		} else {
			switch t := val.(type) {
			case sql.NamedArg:
				val = t.Value
				if t.Name != "" {
					key = t.Name
				}
			case driver.Valuer:
				val, _ = t.Value()
			case time.Time, *time.Time, []byte:
			case fmt.Stringer:
				val = t.String()
			case json.Marshaler:
				jsVal, _ := t.MarshalJSON()
				val = string(jsVal)
			default:
				fieldTypeKind := candihelper.ReflectTypeUnwrapPtr(fieldType.Type).Kind()
				skipSet = fieldTypeKind == reflect.Struct || (fieldTypeKind >= reflect.Array && fieldTypeKind <= reflect.Slice)
			}
		}

		if skipSet && !mustSet {
			continue
		}

		if fieldValue.IsZero() && isDefaultNull {
			val = nil
		}

		if !isPartial {
			updateFields[key] = val
			continue
		}

		_, isFieldUpdated := opt.updateFields[fieldType.Name]
		_, isFieldIgnored := opt.ignoreFields[fieldType.Name]
		if (isFieldUpdated && len(opt.updateFields) > 0) || (!isFieldIgnored && len(opt.ignoreFields) > 0) {
			updateFields[key] = val
		}
	}

	for _, ignored := range d.IgnoredFields {
		delete(updateFields, ignored)
	}
	return updateFields
}

// GetUpdatedFields method
func (d DBUpdateTools) GetFields(opts ...DBUpdateOptionFunc) (updates, ignores []string) {
	opt := d.parseOption(opts...)
	for k := range opt.updateFields {
		updates = append(updates, k)
	}
	for k := range opt.ignoreFields {
		ignores = append(ignores, k)
	}
	return
}
