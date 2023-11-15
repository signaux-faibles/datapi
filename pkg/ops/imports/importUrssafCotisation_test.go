package imports

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_copyFromCotisation(t *testing.T) {
	// given
	ass := assert.New(t)
	tarPath := buildPath(t, "tests/urssafTest.tar.gz")
	tarReader, err := tarFileReader(tarPath)
	require.NoError(t, err)
	for {
		header, err := tarReader.Next()
		ass.NoError(err)
		if strings.HasSuffix(header.Name, "cotisation.csv") {
			break
		}
	}

	cotisationReader := copyFromCotisation{
		reader: tarReader,
	}
	cotisationReader, err = cotisationReader.Init()
	ass.NoError(err)

	// when
	ok := cotisationReader.Next()
	ass.True(ok)
	val, err := cotisationReader.Values()
	ass.NoError(err)

	// then
	ass.Equal("01234567890123", val[0])
	ass.Equal("012345678901234567", val[1])
	ass.Equal(time.Date(1999, time.October, 1, 0, 0, 0, 0, time.UTC), val[2])
}
