package dorm

import (
	"reflect"
	"strconv"
)

// ConvertToType 将reflect.Value的值 暂支持目前这几种
func ConvertToType(value reflect.Value, v interface{}) interface{} {
	kind := value.Kind()
	s, ok := v.(string)
	if !ok {
		return v
	}
	switch kind {
	case reflect.Uint8:
		ret, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return v
		}
		return uint8(ret)
	case reflect.Int:
		ret, err := strconv.Atoi(s)
		if err != nil {
			return v
		}
		return ret
	case reflect.Int32:
		ret, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return v
		}
		return int32(ret)
	case reflect.Int64:
		ret, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return v
		}
		return ret
	default:
		return v
	}
}
