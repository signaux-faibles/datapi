package imports

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_parsePeriode(t *testing.T) {
	// given
	ass := assert.New(t)
	periode := "2000-01-01 00:00:00-2001-01-01 00:00:00"
	expected := [2]time.Time{
		time.Unix(946684800, 0).UTC(),
		time.Unix(978307200, 0).UTC(),
	}
	// when
	actual, _ := parsePeriode(periode)

	// then
	ass.Equal(expected, actual)
}

func Test_parsePeriode_invalidData(t *testing.T) {
	// given
	ass := assert.New(t)
	periode := "2000-13-01 00:00:00-2001-01-01 00:00:00"

	// when
	_, err := parsePeriode(periode)

	// then
	ass.Error(err)
}

func Test_parsePeriode_brokenData(t *testing.T) {
	// given
	ass := assert.New(t)
	periode := "don't read"

	// when
	_, err := parsePeriode(periode)

	// then
	ass.Error(err)
}

func Test_tarFileReader_whenFileDoesntExists(t *testing.T) {
	// given
	ass := assert.New(t)

	// when
	tarReader, err := tarFileReader("tests/nofile.tar.gz")

	// then
	ass.Error(err)
	ass.Nil(tarReader)
}
