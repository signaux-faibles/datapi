package refresh

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"datapi/pkg/test"
)

var tuTime = time.Date(2023, 03, 10, 17, 41, 58, 651387237, time.Local)

func Test_NewRefresh(t *testing.T) {
	test.FakeTime(t, tuTime)
	ass := assert.New(t)
	current := New()
	ass.Equal(Prepare, current.Status)
	ass.NotNil(current.Message)
	ass.Equal(tuTime, current.Date)
	ass.NotNil(current.Date)
	ass.Exactly(*current, last.Load())
	value, found := list.Load(current.UUID)
	ass.True(found)
	ass.Exactly(current, value)
}

func Test_changeRefreshStateToRun(t *testing.T) {
	ass := assert.New(t)
	current := New()
	expectedMessage := "mega test"

	// wait 5"
	expectedTime := tuTime.Add(5 * time.Second)
	test.FakeTime(t, expectedTime)

	current.run(expectedMessage)
	ass.Exactly(expectedMessage, current.Message)
	ass.Exactly(Running, current.Status)
	ass.Exactly(expectedTime, current.Date)
}
