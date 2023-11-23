package imports

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"strings"
	"testing"
)

func Test_copyFromDebit(t *testing.T) {
	// given
	ass := assert.New(t)
	tarReader, err := tarFileReader("tests/urssafTest.tar.gz")
	require.NoError(t, err)
	for {
		header, err := tarReader.Next()
		ass.NoError(err)
		if strings.HasSuffix(header.Name, "debit.csv") {
			break
		}
	}

	debitReader := copyFromDebit{
		reader: tarReader,
	}
	debitReader, err = debitReader.Init()
	ass.NoError(err)

	// when
	debitReader.Next()
	val, err := debitReader.Values()
	ass.NoError(err)

	// then
	ass.Equal("01234567890123", val[0])
	ass.Equal("012345678901234567", val[1])
	ass.Equal("101", val[2])
}
