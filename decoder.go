package dorm

import (
	"errors"
	"github.com/shopspring/decimal"
	"go/ast"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Decoder interface {
	UnmarshalDocument(tagName string, data map[string]interface{}, opt ...interface{}) error
}

func getPrefix(data map[string]interface{}, prefix string) map[string]interface{} {
	if prefix == "" {
		return data
	}
	retMap := map[string]interface{}{}
	for k, v := range data {
		if strings.HasPrefix(k, prefix) {
			nk := strings.TrimPrefix(k, prefix)
			retMap[nk] = v
		}
	}
	return retMap
}

type GroupKey struct {
	Key    string `json:"key"`
	Number string `json:"number"`
}

// 获取分组之后的数据
func getGroupData(data map[string]interface{}, group string) []map[string]interface{} {
	var results []map[string]interface{}
	groupOpts := strings.Split(group, "=")
	if len(groupOpts) <= 0 {
		return results
	}
	groupMethods := groupOpts[0]
	methods := strings.Split(groupMethods, ",")
	groupKeys := map[GroupKey]bool{}
	numbers := map[string]bool{}
	var keys []string
	for _, method := range methods {
		if method == "keys" {
			if len(groupOpts) < 2 {
				return results
			}
			itemStr := groupOpts[1]
			keys = strings.Split(itemStr, " ")
		} else if method == "number" {
			regStr := `[\d]+`
			reg := regexp.MustCompile(regStr)
			for k, _ := range data {
				number := reg.FindString(k)
				if number != "" {
					numbers[number] = true
				}
			}
		}
	}
	if len(keys) > 0 && len(numbers) > 0 {
		for _, key := range keys {
			for number, _ := range numbers {
				gKey := GroupKey{
					Key:    key,
					Number: number,
				}
				groupKeys[gKey] = true
			}
		}
	} else if len(keys) > 0 {
		for _, key := range keys {
			gKey := GroupKey{
				Key: key,
			}
			groupKeys[gKey] = true
		}
	} else if len(numbers) > 0 {
		for number, _ := range numbers {
			gKey := GroupKey{
				Number: number,
			}
			groupKeys[gKey] = true
		}
	}
	regStr := `[\d]+`
	reg := regexp.MustCompile(regStr)
	for groupKey, _ := range groupKeys {
		itemData := map[string]interface{}{}
		for k, v := range data {
			if groupKey.Key != "" && groupKey.Number != "" {
				if strings.Contains(k, groupKey.Key) && reg.FindString(k) == groupKey.Number {
					itemData[k] = v
				}
			} else if groupKey.Key != "" {
				if strings.Contains(k, groupKey.Key) {
					itemData[k] = v
				}
			} else if groupKey.Number != "" {
				if reg.FindString(k) == groupKey.Number {
					itemData[k] = v
				}
			}
		}
		results = append(results, itemData)
	}
	return results
}

// Unmarshal dorm tag 解析
func Unmarshal(v interface{}, data map[string]interface{}, opt ...interface{}) error {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Struct {
		return errors.New("unsupported destination, should be slice or struct")
	}
	reflectType := typ.Elem()
	reflectValue := reflect.ValueOf(v).Elem()
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
					field.Field.Set(reflect.New(field.Struct.Type.Elem()))
				}
				fieldInterface := field.Field.Interface()
				if !isPtr {
					fieldInterface = field.Field.Addr().Interface()
				}
				if name, ok := field.TagSettingsGet("NAME"); ok {
					tagNames := strings.Split(name, "-")
					var prefix string
					if len(tagNames) > 0 {
						prefix = tagNames[0] + "-"
					} else {
						continue
					}
					nData := getPrefix(data, prefix)
					decoder, ok := fieldInterface.(Decoder)
					if ok {
						err := decoder.UnmarshalDocument(name, nData, opt...)
						if err != nil {
							return err
						}
					} else {
						if err := Unmarshal(fieldInterface, nData, opt...); err != nil {
							return err
						}
					}
				}
			case reflect.Slice:
				if name, ok := field.TagSettingsGet("NAME"); ok {
					tagNames := strings.Split(name, "-")
					var prefix string
					if len(tagNames) > 0 {
						prefix = tagNames[0] + "-"
					}
					nData := getPrefix(data, prefix)
					isStruct := field.Field.Type().Elem().Kind() == reflect.Struct
					var groupDataMaps []map[string]interface{}
					if group, ok := field.TagSettingsGet("GROUP"); ok {
						groupDataMaps = getGroupData(nData, group)
					} else {
						for key, val := range nData {
							itemMap := map[string]interface{}{
								key: val,
							}
							groupDataMaps = append(groupDataMaps, itemMap)
						}
					}
					for _, itemMap := range groupDataMaps {
						var elem reflect.Value
						if isStruct {
							elem = reflect.New(field.Field.Type().Elem())
						} else {
							elem = reflect.New(field.Field.Type().Elem()).Elem()
							elem.Set(reflect.New(elem.Type().Elem()))
						}
						itemInterface := elem.Interface()
						decoder, ok := itemInterface.(Decoder)
						if ok {
							err := decoder.UnmarshalDocument(name, itemMap, opt...)
							if err != nil {
								return err
							}
						} else {
							if err := Unmarshal(itemInterface, itemMap, opt...); err != nil {
								return err
							}
						}
						if isStruct {
							field.Field.Set(reflect.Append(field.Field, elem.Elem()))
						} else {
							field.Field.Set(reflect.Append(field.Field, elem))
						}
					}
				}
			default:
				iFace := field.Field.Addr().Interface()
				if decoder, ok := iFace.(Decoder); ok {
					if err := decoder.UnmarshalDocument("", data, opt...); err != nil {
						return err
					}
					continue
				}
				if isReg, ok := field.TagSettingsGet("REG"); ok {
					if isReg == "true" {
						name, ok1 := field.TagSettingsGet("NAME")
						if ok1 {
							reg := regexp.MustCompile(name)
							if reg == nil {
								return errors.New("regexp failed")
							}
							for key, val := range data {
								if reg.MatchString(key) {
									if _, ok := field.TagSettingsGet("FLOAT"); ok {
										if shift, ok := field.TagSettingsGet("SHIFT"); ok {
											if valStr, ok := val.(string); ok {
												floatDeci, err := decimal.NewFromString(valStr)
												if err != nil {
													return err
												}
												shiftVal, err := strconv.Atoi(shift)
												if err != nil {
													return err
												}
												shiftedVal := floatDeci.Shift(int32(shiftVal)).Ceil().IntPart()
												if err := field.Set(shiftedVal); err != nil {
													return err
												}
												continue
											}
											if floatVal, ok := val.(float64); ok {
												floatDeci := decimal.NewFromFloat(floatVal)
												shiftVal, err := strconv.Atoi(shift)
												if err != nil {
													return err
												}
												shiftedVal := floatDeci.Shift(int32(shiftVal)).Ceil().IntPart()
												err = field.Set(shiftedVal)
												return err
											}
											return errors.New("Value cannot convert")
										}
									} else {
										val = ConvertToType(field.Field, val)
										if err := field.Set(val); err != nil {
											return err
										}
									}
								}
							}
						}
					}
				} else {
					if name, ok := field.TagSettingsGet("NAME"); ok {
						val, ok := data[name]
						if ok {
							if _, ok := field.TagSettingsGet("FLOAT"); ok {
								if shift, ok := field.TagSettingsGet("SHIFT"); ok {
									if valStr, ok := val.(string); ok {
										floatDeci, err := decimal.NewFromString(valStr)
										if err != nil {
											return err
										}
										shiftVal, err := strconv.Atoi(shift)
										if err != nil {
											return err
										}
										shiftedVal := floatDeci.Shift(int32(shiftVal)).Ceil().IntPart()
										if err := field.Set(shiftedVal); err != nil {
											return err
										}
										continue
									}
									if floatVal, ok := val.(float64); ok {
										floatDeci := decimal.NewFromFloat(floatVal)
										shiftVal, err := strconv.Atoi(shift)
										if err != nil {
											return err
										}
										shiftedVal := floatDeci.Shift(int32(shiftVal)).Ceil().IntPart()
										err = field.Set(shiftedVal)
										return err
									}
									return errors.New("Value cannot convert")
								}
							}
							val = ConvertToType(field.Field, val)
							if err := field.Set(val); err != nil {
								return err
							}
							if shift, ok := field.TagSettingsGet("SHIFT"); ok {
								if field.Field.Kind() == reflect.Int64 {
									shiftVal, err := strconv.Atoi(shift)
									if err != nil {
										return err
									}
									shiftedVal := decimal.New(field.Field.Int(), int32(shiftVal)).Ceil().IntPart()
									err = field.Set(shiftedVal)
									if err != nil {
										return err
									}
								}
							}
						}
					}
				}
			}

		}
	}
	return nil
}
