package dorm

import (
	"errors"
	"fmt"
)

var (
	ErrUnaddressable = errors.New("using unaddressable value")
)

type RowError struct {
	MetaInfo interface{}
	Errors    string
}

func NewRowError(metaInfo interface{}, text string) error {
	return &RowError{
		MetaInfo: metaInfo,
		Errors:   text,
	}
}

func (e *RowError) Error() string {
	return fmt.Sprintf("%v, error:%s", e.MetaInfo, e.Errors)
}
