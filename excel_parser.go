package dorm

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// ExcelParser excel解析器 实现了Parser接口
type ExcelParser struct {
	sheetName string
	file      *excelize.File
}

// MetaInfo excel的元信息
type MetaInfo struct {
	SheetName string
	RowIndex  string
}

// ExcelRow excel行信息
type ExcelRow struct {
	data     map[string]interface{}
	metaInfo MetaInfo
}

// GetData 获取到当前行的数据
func (m *ExcelRow) GetData() map[string]interface{} {
	return m.data
}

// GetMetaInfo 获取当前元信息
func (m *ExcelRow) GetMetaInfo() interface{} {
	return m.metaInfo
}

// NewExcelParser 通过文件reader实例化一个ExcelParser
func NewExcelParser(r io.Reader) (*ExcelParser, error) {
	xlsFile, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	parser := &ExcelParser{file: xlsFile}
	return parser, nil
}

// SetSheetName 设置当前sheetName
func (p *ExcelParser) SetSheetName(sheetName string) {
	p.sheetName = sheetName
}

func (p *ExcelParser) rowsToResults(sheetName string, columns [][]string, rows []RowInterface) []RowInterface {
	titleIndex := map[int]string{}
	for index, column := range columns {
		if index == 0 {
			for rowIndex, colCell := range column {
				titleIndex[rowIndex] = strings.TrimSpace(colCell)
			}
		} else {
			rowData := map[string]interface{}{}
			for rowIndex, colCell := range column {
				titleName := titleIndex[rowIndex]
				rowData[titleName] = colCell
			}
			cellResult := &ExcelRow{
				data: rowData,
				metaInfo: MetaInfo{
					SheetName: sheetName,
					RowIndex:  strconv.Itoa(index),
				},
			}
			rows = append(rows, cellResult)
		}
	}
	return rows
}

// ReadToRows 读取并解析道行数据列表
func (p *ExcelParser) ReadToRows(opt ...interface{}) ([]RowInterface, error) {
	var rows []RowInterface
	if p.file.SheetCount <= 0 {
		return nil, errors.New("SheetCount is zero")
	}
	var columns [][]string
	if p.sheetName != "" {
		columns = p.file.GetRows(p.sheetName)
		rows = p.rowsToResults(p.sheetName, columns, rows)
	} else {
		for i := 1; i <= p.file.SheetCount; i++ {
			sheetName := p.file.GetSheetName(i)
			sheetRows := p.file.GetRows(sheetName)
			rows = p.rowsToResults(sheetName, sheetRows, rows)
		}
	}
	return rows, nil
}
