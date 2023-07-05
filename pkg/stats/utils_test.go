package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_transformCompressLogsToData(t *testing.T) {
	ass := assert.New(t)
	input := []line{
		{
			date:     time.Date(2023, 12, 24, 1, 2, 3, 4, time.UTC),
			path:     "some/path",
			method:   "GET",
			username: "gounit",
			roles: []string{
				"role1", "role2",
			},
		},
	}
	_, err := transformLogsToCompressedData(input)
	ass.NoError(err)
}
