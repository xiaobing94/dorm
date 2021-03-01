package dorm

import (
	"go/ast"
	"reflect"
)

type Encoder interface {
	MarshalDocument() (map[string]interface{}, error)
}

// Marshal dorm tag 解析
func Marshal(v interface{}) (map[string]interface{}, error) {
	typ := reflect.TypeOf(v)
	reflectType := typ.Elem()
	reflectValue := reflect.ValueOf(v).Elem()
	result := map[string]interface{}{}
	for i := 0; i < reflectType.NumField(); i++ {
		if fieldStruct := reflectType.Field(i); ast.IsExported(fieldStruct.Name) {
			kind := reflectValue.Field(i).Kind()
			field := &Field{
				StructField: &StructField{
					Struct:      fieldStruct,
					Name:        fieldStruct.Name,
					Tag:         fieldStruct.Tag,
					TagSettings: parseTagSetting(fieldStruct.Tag),
				},
				Field: reflectValue.Field(i),
			}
			// is ignored field
			if _, ok := field.TagSettingsGet("-"); ok {
				continue
			}
			switch kind {
			case reflect.Ptr, reflect.Struct:
				isPtr := kind == reflect.Ptr
				if isPtr && field.Field.IsNil() {
					continue
				}
				fieldInterface := field.Field.Interface()
				if !isPtr {
					fieldInterface = field.Field.Addr().Interface()
				}
				var subResult map[string]interface{}
				var err error
				encoder, ok := fieldInterface.(Encoder)
				if ok {
					subResult, err = encoder.MarshalDocument()
					if err != nil {
						return nil, err
					}
				} else {
					subResult, err = Marshal(fieldInterface)
					if err != nil {
						return nil, err
					}
				}
				if name, ok := field.TagSettingsGet("NAME"); ok {
					for key, val := range subResult {
						result[name+key] = val
					}
				}
			default:
				if name, ok := field.TagSettingsGet("NAME"); ok {
					result[name] = field.Field.Interface()
				}
			}
		}
	}
	return result, nil
}
