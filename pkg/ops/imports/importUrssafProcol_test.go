package imports

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"strings"
	"testing"
	"time"
)

func Test_copyFromProcol(t *testing.T) {
	// given
	ass := assert.New(t)
	abs := buildPath(t, "tests/urssafTest.tar.gz")
	tarReader, err := tarFileReader(abs)
	require.NoError(t, err)
	for {
		header, err := tarReader.Next()
		ass.NoError(err)
		if strings.HasSuffix(header.Name, "procol.csv") {
			break
		}
	}

	procolReader := copyFromProcol{
		reader: tarReader,
	}
	procolReader, err = procolReader.Init()
	ass.NoError(err)

	// when
	ok := procolReader.Next()
	ass.True(ok)
	val, err := procolReader.Values()
	ass.NoError(err)

	// then
	ass.Equal("01234567890123", val[0])
	ass.Equal(time.Date(2017, time.October, 25, 0, 0, 0, 0, time.UTC), val[1])
	ass.Equal("redressement", val[2])
	ass.Equal("ouverture", val[3])
}
