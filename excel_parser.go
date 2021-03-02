package dorm

import (
	"errors"
	"github.com/360EntSecGroup-Skylar/excelize"
	"io"
	"strconv"
	"strings"
)

type ExcelParser struct {
	sheetName string
	file      *excelize.File
}

type MetaInfo struct {
	SheetName string
	RowIndex  string
}

type CellResult struct {
	result   map[string]interface{}
	metaInfo MetaInfo
}

func (m *CellResult) GetResults() map[string]interface{} {
	return m.result
}

func (m *CellResult) GetMetaInfo() interface{} {
	return m.metaInfo
}

func NewExcelParser(r io.Reader) (*ExcelParser, error) {
	xlsFile, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	parser := &ExcelParser{file: xlsFile}
	return parser, nil
}

func (p *ExcelParser) SetSheetName(sheetName string) {
	p.sheetName = sheetName
}

func (p *ExcelParser) rowsToResults(sheetName string, rows [][]string, results []ResultInterface) []ResultInterface {
	titleIndex := map[int]string{}
	for index, row := range rows {
		if index == 0 {
			for rowIndex, colCell := range row {
				titleIndex[rowIndex] = strings.TrimSpace(colCell)
			}
		} else {
			cell := map[string]interface{}{}
			for rowIndex, colCell := range row {
				titleName := titleIndex[rowIndex]
				cell[titleName] = colCell
			}
			cellResult := &CellResult{
				result: cell,
				metaInfo: MetaInfo{
					SheetName: sheetName,
					RowIndex:  strconv.Itoa(index),
				},
			}
			results = append(results, cellResult)
		}
	}
	return results
}

func (p *ExcelParser) ReadToMaps(opt ...interface{}) ([]ResultInterface, error) {
	var results []ResultInterface
	if p.file.SheetCount <= 0 {
		return nil, errors.New("SheetCount is zero")
	}
	var rows [][]string
	if p.sheetName != "" {
		rows = p.file.GetRows(p.sheetName)
		results = p.rowsToResults(p.sheetName, rows, results)
	} else {
		for i := 1; i < p.file.SheetCount+1; i++ {
			sheetName := p.file.GetSheetName(i)
			sheetRows := p.file.GetRows(sheetName)
			results = p.rowsToResults(sheetName, sheetRows, results)
		}
	}
	return results, nil
}
