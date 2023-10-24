package stats

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
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
	xls := newExcel()
	sheetName := fmt.Sprintf("%.25s", fakeTU.Lorem().Sentence(3))
	item1 := fakeTU.Lorem().Word()
	item2 := fakeTU.Lorem().Word()
	itemsToWrite := createStringChannel(item1, item2)
	sheetConf := dummySheetConfig(sheetName)
	err := writeOneSheetToExcel2(xls, sheetConf, itemsToWrite)
	require.NoError(t, err)
	assert.Len(t, xls.WorkBook.Sheets.Sheet, 1)
	assert.Equal(t, xls.WorkBook.Sheets.Sheet[0].Name, sheetName)
	assert.Contains(t, xls.GetSheetList(), sheetName)

	// header
	header := sheetConf.headers()[0]
	assertCellValue(t, xls, sheetName, 1, 1, header)

	// interligne
	assertCellValue(t, xls, sheetName, 1, 2, "")

	// value 1
	assertCellValue(t, xls, sheetName, 1, 3, item1)

	// value 2
	assertCellValue(t, xls, sheetName, 1, 4, item2)
}

func assertCellValue(t *testing.T, xls *excelize.File, sheetName string, col, row int, expected any) {
	cell, err := excelize.CoordinatesToCellName(col, row)
	require.NoError(t, err)
	actual, err := xls.GetCellValue(sheetName, cell)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func dummySheetConfig(label string) sheetConfig[dummy] {
	toRow := func(s dummy) []any {
		return []any{s.value}
	}
	return anySheetConfig[dummy]{
		item:      dummy{},
		sheetName: label,
		startRow:  3,
		asRow:     toRow,
	}
}

func Test_writeSheetToExcel(t *testing.T) {

	output, err := os.CreateTemp(os.TempDir(), t.Name()+"_*.xls")
	require.NoError(t, err)
	t.Log("destination du fichier excel", output.Name())
	xls := newExcel()
	sheetName := fakeTU.Lorem().Sentence(3)
	_, err = addSheet(xls, sheetName)
	require.NoError(t, err)
	err = xls.SetSheetRow(sheetName, "B6", &[]interface{}{"1", nil, 2})
	require.NoError(t, err)

	err = xls.Write(output)
	require.NoError(t, err)
}
