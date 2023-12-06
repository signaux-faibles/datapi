package stats

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"datapi/pkg/test"
)

func Test_addSheet(t *testing.T) {
	excel := newExcel()
	idx1, err := addSheet(excel, "feuille 1")
	require.NoError(t, err)
	idx2, err := addSheet(excel, "feuille 2")
	require.NoError(t, err)
	assert.Equal(t, 0, idx1)
	assert.Equal(t, "feuille 1", excel.GetSheetName(0))
	assert.Len(t, excel.GetSheetList(), 2)
	assert.NotEqual(t, "Sheet 1", excel.GetSheetName(0))
	assert.Equal(t, 1, idx2)
	assert.Equal(t, "feuille 2", excel.GetSheetName(1))
}

func Test_writeOneSheetToExcel(t *testing.T) {
	output := test.CreateTempExcelFile(t, t.Name())
	xls := newExcel()

	// GIVEN
	sheetName := randomSheetName()
	item1 := createTestLine()
	item2 := createTestLine()
	itemsToWrite := createTestChannel(item1, item2)
	sheetConf := dummySheetConfig(sheetName)

	// WHEN
	err := writeOneSheetToExcel(xls, sheetConf, itemsToWrite)
	require.NoError(t, err)

	// THEN
	assert.Len(t, xls.WorkBook.Sheets.Sheet, 1)
	assert.Equal(t, xls.WorkBook.Sheets.Sheet[0].Name, sheetName)
	assert.Contains(t, xls.GetSheetList(), sheetName)

	// header
	headers, err := sheetConf.headers()
	require.NoError(t, err)
	assertCellValue(t, xls, sheetName, 1, 1, headers[0])

	// value 1
	assertCellValue(t, xls, sheetName, 1, 2, item1.valString)
	assertCellValue(t, xls, sheetName, 2, 2, item1.valDay.Format(time.DateOnly))
	assertCellValue(t, xls, sheetName, 3, 2, item1.valInstant.Format(time.DateTime))
	assertCellValue(t, xls, sheetName, 4, 2, fmt.Sprint(item1.valInt))

	// value 2
	// assertCellValue(t, xls, sheetName, 1, 3, item2)

	err = xls.Write(output)
	require.NoError(t, err)
}

func Test_headers(t *testing.T) {
	sheetName := randomSheetName()
	config := dummySheetConfig(sheetName)

	actual, _ := config.headers()
	assert.ElementsMatch(t, actual, []string{"chaîne de caractères", "jour", "instant", "entier"})
}

func assertCellValue(t *testing.T, xls *excelize.File, sheetName string, col, row int, expected any) {
	cell, err := excelize.CoordinatesToCellName(col, row)
	require.NoError(t, err)
	actual, err := xls.GetCellValue(sheetName, cell)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func dummySheetConfig(label string) sheetConfig[testline] {
	toRow := func(s testline) []any {
		return []any{s.valString, s.valDay, s.valInstant, s.valInt}
	}
	return anySheetConfig[testline]{
		item:      testline{},
		sheetName: label,
		asRow:     toRow,
		mapStyles: map[int]excelize.Style{1: styleDateOnly, 2: styleDateTime},
	}
}

func randomSheetName() string {
	sentence := fakeTU.Lorem().Sentence(3)
	if len(sentence) <= 31 {
		return sentence
	}
	return sentence[:31]
}
