package imports

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func Test_copyFromEffectif(t *testing.T) {
	// given
	ass := assert.New(t)
	tarReader, err := tarFileReader("tests/urssafTest.tar.gz")
	for {
		header, err := tarReader.Next()
		ass.NoError(err)
		if strings.HasSuffix(header.Name, "effectif.csv") {
			break
		}
	}

	effectifReader := copyFromEffectif{
		reader: tarReader,
	}
	effectifReader, err = effectifReader.Init()
	ass.NoError(err)

	// when
	ok := effectifReader.Next()
	ass.True(ok)
	val, err := effectifReader.Values()
	ass.NoError(err)

	// then
	ass.Equal("01234567890123", val[0])
	ass.Equal("012345678901234567", val[1])
	ass.Equal(time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC), val[2])
}
