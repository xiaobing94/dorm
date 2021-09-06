package dorm

import (
	"errors"
	"io"
	"os"
	"reflect"
)

// RowInterface 行数据接口
type RowInterface interface {
	// GetData 获取到当前行的数据
	GetData() map[string]interface{}
	// GetMetaInfo 获取当前元信息
	GetMetaInfo() interface{}
}

// Parser 解析到为结果的方法 MetaInfo 信息为自定义
type Parser interface {
	// ReadToRows 读取并解析道行数据列表
	ReadToRows(opt ...interface{}) ([]RowInterface, error)
}

type ObjectMapper interface {
	// SetParser 设置一个解析器
	SetParser(parser Parser)
	// Encode 将文档编码为指定的对象
	Encode(v interface{}, opt ...interface{}) error
}

// DocumentMapper 文档映射转换工具
type DocumentMapper struct {
	errs   []error
	parser Parser
}

// SetParser 设置一个解析器
func (mapper *DocumentMapper) SetParser(parser Parser) {
	mapper.parser = parser
}

// GetErrors 获取解析中发生的错误
func (mapper *DocumentMapper) GetErrors() []error {
	return mapper.errs
}

// Encode 将文档编码为指定的对象
func (mapper *DocumentMapper) Encode(v interface{}, opt ...interface{}) error {
	errs, err := EncodeByParser(mapper.parser, v, opt...)
	mapper.errs = errs
	return err
}

// OpenXlsFile 打开excel文件并使用默认的excel解析方式
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

// OpenReader 打开一个reader对象,并使用默认的excel解析方法
func OpenReader(file io.Reader) (*DocumentMapper, error) {
	parser, err := NewExcelParser(file)
	if err != nil {
		return nil, err
	}
	mapper := &DocumentMapper{parser: parser}
	return mapper, nil
}

// EncodeByParser 使用parser解析文档为对象
func EncodeByParser(parser Parser, v interface{}, opt ...interface{}) ([]error, error) {
	var errs []error
	value := reflect.ValueOf(v)
	if !value.IsValid() {
		return nil, errors.New("interface not valid")
	}

	kind := reflect.TypeOf(v).Kind()
	if kind != reflect.Ptr {
		return nil, errors.New("v must be ptr")
	}
	results, err := parser.ReadToRows()
	if err != nil {
		return nil, err
	}
	reflectValue := reflect.ValueOf(v).Elem()
	if reflectValue.Kind() != reflect.Slice {
		return nil, errors.New("v mast be []*T type")
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
		if err := encodeResult(itemInterface, r, opt...); err != nil {
			errs = append(errs, WrapError(r.GetMetaInfo(), err))
			continue
		}
		if isStruct {
			reflectValue.Set(reflect.Append(reflectValue, elem.Elem()))
		} else {
			reflectValue.Set(reflect.Append(reflectValue, elem))
		}
	}
	return errs, nil
}

func encodeResult(v interface{}, rowInterface RowInterface, opt ...interface{}) error {
	data := rowInterface.GetData()
	row := &Row{
		Data:     data,
		MetaInfo: rowInterface.GetMetaInfo(),
	}
	if encoder, ok := v.(Encoder); ok {
		if err := encoder.EncodeDocument(row, opt...); err != nil {
			return err
		}
	} else {
		if err := Encode(v, row, opt...); err != nil {
			return err
		}
	}
	return nil
}

// WriteToExcelFile 将对象写入到excel的指定的sheet中
func WriteToExcelFile(writer io.Writer, sheetName string, v interface{}) error {
	var err error
	var nameValues []map[string]interface{}
	var nameValue map[string]interface{}
	var nameSorts []string

	value := reflect.ValueOf(v)
	if !value.IsValid() {
		return errors.New("interface not valid")
	}

	kind := reflect.TypeOf(v).Kind()
	if kind != reflect.Slice {
		return errors.New("interface must be slice")
	}
	reflectValue := reflect.ValueOf(v)
	length := reflectValue.Len()
	for i := 0; i < length; i++ {
		value := reflectValue.Index(i)
		vi := value.Interface()
		decoder, ok := vi.(Decoder)
		if ok {
			nameValue, nameSorts, err = decoder.DecodeDocument()
			if err != nil {
				return err
			}
		} else {
			nameValue, nameSorts, err = DecodeDocument(vi)
			if err != nil {
				return err
			}
		}
		nameValues = append(nameValues, nameValue)
	}

	titles := map[string]interface{}{}
	for _, name := range nameSorts {
		titles[name] = name
	}

	var titleValues []map[string]interface{}
	if len(titles) > 0 {
		titleValues = append(titleValues, titles)
	}

	serializer := NewExcelSerializer()
	serializer.Serialize(nameSorts, 1, sheetName, titleValues)
	serializer.Serialize(nameSorts, 2, sheetName, nameValues)
	_, err = serializer.WriteToFile(writer)
	return err
}
