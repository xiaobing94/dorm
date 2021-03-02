package dorm

import (
	"errors"
	"io"
	"os"
	"reflect"
)

type ResultInterface interface {
	GetResults() map[string]interface{}
	GetMetaInfo() interface{}
}

type Parser interface {
	ReadToMaps(opt ...interface{}) ([]ResultInterface, error)
}

type ObjectMapper interface {
	SetParser(parser Parser)
	GetObjectsFromParser(v interface{}) error
}

type DocumentMapper struct {
	parser Parser
}

func (mapper *DocumentMapper) SetParser(parser Parser) {
	mapper.parser = parser
}

func (mapper *DocumentMapper) GetObjectsFromParser(v interface{}, opt ...interface{}) error {
	return GetObjectsFromParser(mapper.parser, v, opt...)
}

func OpenXlsFile(filename string) (*DocumentMapper, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	parser, err := NewExcelParser(file)
	if err != nil {
		return nil, err
	}
	mapper := &DocumentMapper{parser: parser}
	return mapper, nil
}

func OpenReader(file io.Reader) (*DocumentMapper, error) {
	parser, err := NewExcelParser(file)
	if err != nil {
		return nil, err
	}
	mapper := &DocumentMapper{parser: parser}
	return mapper, nil
}

func GetObjectsFromParser(parser Parser, v interface{}, opt ...interface{}) error {
	value := reflect.ValueOf(v)
	if !value.IsValid() {
		return errors.New("interface not valid")
	}

	kind := reflect.TypeOf(v).Kind()
	if kind != reflect.Ptr {
		return errors.New("v must be ptr")
	}
	results, err := parser.ReadToMaps()
	if err != nil {
		return err
	}
	reflectValue := reflect.ValueOf(v).Elem()
	if reflectValue.Kind() != reflect.Slice {
		return errors.New("v mast be []*T type")
	}
	var elem reflect.Value
	isStruct := reflectValue.Type().Elem().Kind() == reflect.Struct
	for _, r := range results {
		if isStruct {
			elem = reflect.New(reflectValue.Type().Elem())
		} else {
			elem = reflect.New(reflectValue.Type().Elem()).Elem()
			elem.Set(reflect.New(elem.Type().Elem()))
		}
		itemInterface := elem.Interface()
		if decoder, ok := itemInterface.(Decoder); ok {
			m := r.GetResults()
			if err := decoder.UnmarshalDocument("", m, r.GetMetaInfo(), opt...); err != nil {
				return err
			}
		} else {
			m := r.GetResults()
			if err := Unmarshal(itemInterface, m, r.GetMetaInfo(), opt...); err != nil {
				return err
			}
		}
		if isStruct {
			reflectValue.Set(reflect.Append(reflectValue, elem.Elem()))
		} else {
			reflectValue.Set(reflect.Append(reflectValue, elem))
		}
	}
	return nil
}
