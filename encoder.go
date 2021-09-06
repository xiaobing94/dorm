package dorm

import (
	"go/ast"
	"reflect"
	"sort"
)

// Decoder 解码器
type Decoder interface {
	DecodeDocument() (map[string]interface{}, []string, error)
}

// WeightKey 带权重的key
type WeightKey struct {
	Key    string
	weight string
}

// WeightKeys 带权重key列表
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

// GetKeys 获取所有的key
func (k WeightKeys) GetKeys() []string {
	var keys []string
	for _, weightKey := range k {
		keys = append(keys, weightKey.Key)
	}
	return keys
}

// DecodeDocument 解码对象到map
func DecodeDocument(v interface{}) (map[string]interface{}, []string, error) {
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
				var err error
				result, err = decodeDecodeDocumentStruct(kind, field, result)
				if err != nil {
					return nil, nil, err
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

func decodeDecodeDocumentStruct(kind reflect.Kind, field *Field,
	result map[string]interface{}) (map[string]interface{}, error){
	isPtr := kind == reflect.Ptr
	if isPtr && field.Field.IsNil() {
		return result, nil
	}
	fieldInterface := field.Field.Interface()
	if !isPtr {
		fieldInterface = field.Field.Addr().Interface()
	}
	var subResult map[string]interface{}
	var err error
	decoder, ok := fieldInterface.(Decoder)
	if ok {
		subResult, _, err = decoder.DecodeDocument()
		if err != nil {
			return result, err
		}
	} else {
		subResult, _, err = DecodeDocument(fieldInterface)
		if err != nil {
			return result, err
		}
	}
	if name, ok := field.TagSettingsGet("NAME"); ok {
		for key, val := range subResult {
			result[name+key] = val
		}
	}
	return result, nil
}