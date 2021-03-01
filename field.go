package dorm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type StructField struct {
	Name            string
	Tag             reflect.StructTag
	TagSettings     map[string]string
	Struct          reflect.StructField
	tagSettingsLock sync.RWMutex
}

// TagSettingsSet
func (sf *StructField) TagSettingsSet(key, val string) {
	sf.tagSettingsLock.Lock()
	defer sf.tagSettingsLock.Unlock()
	sf.TagSettings[key] = val
}

// TagSettingsGet
func (sf *StructField) TagSettingsGet(key string) (string, bool) {
	sf.tagSettingsLock.RLock()
	defer sf.tagSettingsLock.RUnlock()
	val, ok := sf.TagSettings[key]
	return val, ok
}

// TagSettingsDelete
func (sf *StructField) TagSettingsDelete(key string) {
	sf.tagSettingsLock.Lock()
	defer sf.tagSettingsLock.Unlock()
	delete(sf.TagSettings, key)
}

type Field struct {
	*StructField
	Field reflect.Value
}

// Set set a value to the field
func (field *Field) Set(value interface{}) (err error) {
	if !field.Field.IsValid() {
		return errors.New("field value not valid")
	}

	if !field.Field.CanAddr() {
		return ErrUnaddressable
	}

	reflectValue, ok := value.(reflect.Value)
	if !ok {
		reflectValue = reflect.ValueOf(value)
	}

	fieldValue := field.Field
	if reflectValue.IsValid() {
		if reflectValue.Type().ConvertibleTo(fieldValue.Type()) {
			fieldValue.Set(reflectValue.Convert(fieldValue.Type()))
		} else {
			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					fieldValue.Set(reflect.New(field.Struct.Type.Elem()))
				}
				fieldValue = fieldValue.Elem()
			}

			if reflectValue.Type().ConvertibleTo(fieldValue.Type()) {
				fieldValue.Set(reflectValue.Convert(fieldValue.Type()))
			} else {
				err = fmt.Errorf("could not convert argument of field %s from %s to %s", field.Name, reflectValue.Type(), fieldValue.Type())
			}
		}
	} else {
		field.Field.Set(reflect.Zero(field.Field.Type()))
	}

	return err
}

func parseTagSetting(tags reflect.StructTag) map[string]string {
	setting := map[string]string{}
	for _, str := range []string{tags.Get("dorm")} {
		if str == "" {
			continue
		}
		tags := strings.Split(str, ";")
		for _, value := range tags {
			v := strings.Split(value, ":")
			k := strings.TrimSpace(strings.ToUpper(v[0]))
			if len(v) >= 2 {
				setting[k] = strings.Join(v[1:], ":")
			} else {
				setting[k] = k
			}
		}
	}
	return setting
}