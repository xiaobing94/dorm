package dorm

import (
	"go/ast"
	"reflect"
	"sort"
)

type Encoder interface {
	MarshalDocument() (map[string]interface{}, []string, error)
}

type WeightKey struct {
	Key    string
	weight string
}

type WeightKeys []WeightKey

func (k WeightKeys) Len() int {
	return len(k)
}

func (k WeightKeys) Less(i, j int) bool {
	if k[i].weight == "" && k[j].weight != "" {
		return false
	}
	if k[i].weight != "" && k[j].weight == "" {
		return true
	}
	return k[i].weight < k[j].weight
}

func (k WeightKeys) Swap(i, j int) {
	k[i].weight, k[j].weight = k[j].weight, k[i].weight
	k[i].Key, k[j].Key = k[j].Key, k[i].Key
}

func (k WeightKeys) GetKeys() []string {
	var keys []string
	for _, weightKey := range k {
		keys = append(keys, weightKey.Key)
	}
	return keys
}

// Marshal dorm tag 解析
func Marshal(v interface{}) (map[string]interface{}, []string, error) {
	var keySort []string
	var weightKeys WeightKeys
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
					subResult, _, err = encoder.MarshalDocument()
					if err != nil {
						return nil, nil, err
					}
				} else {
					subResult, _, err = Marshal(fieldInterface)
					if err != nil {
						return nil, nil, err
					}
				}
				if name, ok := field.TagSettingsGet("NAME"); ok {
					for key, val := range subResult {
						result[name+key] = val
					}
				}
			default:
				if name, ok := field.TagSettingsGet("NAME"); ok {
					weight, _ := field.TagSettingsGet("INDEX")
					weightKeys = append(weightKeys, WeightKey{
						Key:    name,
						weight: weight,
					})
					result[name] = field.Field.Interface()
				}
			}
		}
	}
	if len(weightKeys) > 0 {
		sort.Sort(weightKeys)
		keySort = weightKeys.GetKeys()
	}
	return result, keySort, nil
}
