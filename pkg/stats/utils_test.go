package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_transformLogsToData(t *testing.T) {
	ass := assert.New(t)
	input := []line{
		{
			date:     time.Now(),
			path:     "some/path",
			method:   "GET",
			username: "gounit",
			roles: []string{
				"role1", "role2",
			},
		},
	}
	data, err := transformLogsToData(input)
	t.Logf("line -> %s", data)
	ass.NoError(err)
	ass.Less(0, len(data))
}
