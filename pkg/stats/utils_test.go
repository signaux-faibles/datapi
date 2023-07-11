package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"datapi/pkg/test"
)

func Test_transformCompressLogsToData(t *testing.T) {
	ass := assert.New(t)
	results := make(chan accessLog)
	go writeLines(results, 999) // ==> le résultat compressé est 233 bytes
	data, err := writeLinesToCompressedCSV(results, 999999)
	ass.NoError(err)
	ass.Len(data, 233)
	lines, err := test.ReadGZippedCSV(data)
	ass.NoError(err)
	ass.Len(lines, 999)
}

func writeLines(results chan accessLog, number int) {
	for i := 0; i < number; i++ {
		results <- accessLog{
			date:     time.Date(2023, 12, 24, 1, 2, 3, 4, time.UTC),
			path:     "some/path",
			method:   "GET",
			username: "gounit",
			roles: []string{
				"role1", "role2",
			},
		}
	}
	close(results)
}
