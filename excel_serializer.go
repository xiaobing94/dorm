package dorm

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"io"
)

type ExcelSerializer struct {
	excelFile *excelize.File
}

func NewExcelSerializer() *ExcelSerializer {
	excelFile := excelize.NewFile()
	serializer := &ExcelSerializer{
		excelFile:excelFile,
	}
	return serializer
}

func (es *ExcelSerializer) Serialize(keys []string, startRow int, data []map[string]interface{}) {
	if len(data) <= 0 {
		return
	}
	if len(keys) <= 0 {
		for key, _ := range data[0] {
			keys = append(keys, key)
		}
	}
	sheet := "Sheet1"
	for index, m := range data {
		for ki, key := range keys {
			axis := excelize.ToAlphaString(ki) + fmt.Sprintf("%d", startRow+index)
			if v, ok := m[key]; ok {
				es.excelFile.SetCellValue(sheet, axis, v)
			}
		}
	}
	return
}

func (es *ExcelSerializer) WriteToFile(writer io.Writer) (int64, error) {
	return es.excelFile.WriteTo(writer)
}