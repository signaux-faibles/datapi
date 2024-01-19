package imports

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseScores(t *testing.T) {
	ass := assert.New(t)

	// given
	filename := "tests/test_liste.json"

	// when
	scores, err := parseScores(filename)
	ass.NoError(err)

	// then
	ass.Len(scores, 2)
}

func TestCopyFromScore(t *testing.T) {
	ass := assert.New(t)

	// étant donné les éléments suivants
	filename := "tests/test_liste.json"

	copyFromScore := CopyFromScore{}
	copyFromScore.init(filename, "2312", "test")

	// lorsqu'on exécute
	values, err := copyFromScore.Values()
	ass.NoError(err)

	// alors il doit se produire
	ass.NotNil(values)
}
