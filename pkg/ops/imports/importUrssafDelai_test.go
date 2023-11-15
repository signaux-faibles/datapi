package imports

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_copyFromDelai(t *testing.T) {
	// given
	ass := assert.New(t)
	tarPath := buildPath(t, "tests/urssafTest.tar.gz")
	tarReader, err := tarFileReader(tarPath)
	require.NoError(t, err)
	for {
		header, err := tarReader.Next()
		ass.NoError(err)
		if strings.HasSuffix(header.Name, "delai.csv") {
			break
		}
	}

	delaiReader := copyFromDelai{
		reader: tarReader,
	}
	delaiReader, err = delaiReader.Init()
	ass.NoError(err)

	// when
	delaiReader.Next()
	val, err := delaiReader.Values()
	ass.NoError(err)

	// then
	ass.Equal("01234567890123", val[0])
	ass.Equal("012345678901234567", val[1])
	ass.Equal("0123456789012", val[2])
}
