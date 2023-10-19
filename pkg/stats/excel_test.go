package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
