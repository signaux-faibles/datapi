package stats

import (
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
	sheetName := fakeTU.Lorem().Sentence(3)
	item1 := fakeTU.Lorem().Word()
	item2 := fakeTU.Lorem().Word()
	itemsToWrite := createSomeItemsToWrite(item1, item2)
	err := writeOneSheetToExcel(xls, sheetName, itemsToWrite, stringWriter)
	require.NoError(t, err)
	assert.Len(t, xls.WorkBook.Sheets.Sheet, 1)
	assert.Equal(t, xls.WorkBook.Sheets.Sheet[0].Name, sheetName)
	assert.Contains(t, xls.GetSheetList(), sheetName)

	name, err := excelize.CoordinatesToCellName(1, 1)
	require.NoError(t, err)
	actual, err := xls.GetCellValue(sheetName, name)
	require.NoError(t, err)
	assert.Equal(t, item1, actual)

	name, err = excelize.CoordinatesToCellName(1, 2)
	require.NoError(t, err)
	actual, err = xls.GetCellValue(sheetName, name)
	require.NoError(t, err)
	assert.Equal(t, item2, actual)
}

func stringWriter(f *excelize.File, name string, ligne string, r int) error {
	return writeString(f, name, ligne, 1, r)
}

func createSomeItemsToWrite(items ...string) chan row[string] {
	r := make(chan row[string])
	go func() {
		defer close(r)
		for _, i := range items {
			r <- newRow(i)
		}
	}()
	return r
}
