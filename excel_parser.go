package dorm

import (
	"errors"
	"github.com/360EntSecGroup-Skylar/excelize"
	"io"
)

type ExcelParser struct {
	sheetName string
	file      *excelize.File
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

func (p *ExcelParser) ReadToMaps(opt ...interface{}) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	titleIndex := map[int]string{}
	if p.file.SheetCount <= 0 {
		return nil, errors.New("SheetCount is zero")
	}
	var rows [][]string
	if p.sheetName != "" {
		rows = p.file.GetRows(p.sheetName)
	} else {
		for i := 1; i < p.file.SheetCount+1; i++ {
			sheetName := p.file.GetSheetName(i)
			sheetRows := p.file.GetRows(sheetName)
			if i > 1 {
				sheetRows = sheetRows[1:]
			}
			rows = append(rows, sheetRows...)
		}
	}

	for index, row := range rows {
		if index == 0 {
			for rowIndex, colCell := range row {
				titleIndex[rowIndex] = colCell
			}
		} else {
			tmp := map[string]interface{}{}
			for rowIndex, colCell := range row {
				titleName := titleIndex[rowIndex]
				tmp[titleName] = colCell
			}
			results = append(results, tmp)
		}
	}
	return results, nil
}
