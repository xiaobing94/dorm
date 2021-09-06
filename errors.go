package dorm

import (
	"errors"
	"fmt"
)

var (
	ErrUnaddressable = errors.New("using unaddressable value")
)

type RowError struct {
	MetaInfo  interface{}
	ErrorInfo string
}

func NewRowError(metaInfo interface{}, text string) error {
	return &RowError{
		MetaInfo:  metaInfo,
		ErrorInfo: text,
	}
}

func WrapError(metaInfo interface{}, err error) error {
	return &RowError{
		MetaInfo:  metaInfo,
		ErrorInfo: err.Error(),
	}
}

func (e *RowError) Error() string {
	return fmt.Sprintf("%v, error:%s", e.MetaInfo, e.ErrorInfo)
}
