package refresh

import (
	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var TuTime = time.Date(2023, 03, 10, 17, 41, 58, 651387237, time.UTC)

func fakeTime(t time.Time) {
	monkey.Patch(time.Now, func() time.Time {
		return t
	})
}

func TestMain(m *testing.M) {
	fakeTime(TuTime)
	m.Run()
}

func Test_NewRefresh(t *testing.T) {
	ass := assert.New(t)
	current := New()
	ass.Equal(Prepare, current.Status)
	ass.NotNil(current.Message)
	ass.Equal(TuTime, current.Date)
	ass.NotNil(current.Date)
	ass.Exactly(*current, last.Load())
	value, found := list.Load(current.UUID)
	ass.True(found)
	ass.Exactly(current, value)
}

func Test_Run(t *testing.T) {
	ass := assert.New(t)
	current := New()
	expectedMessage := "mega test"

	// wait 5"
	expectedTime := TuTime.Add(5 * time.Second)
	fakeTime(expectedTime)

	current.run(expectedMessage)
	ass.Exactly(expectedMessage, current.Message)
	ass.Exactly(Running, current.Status)
	ass.Exactly(TuTime.Add(5*time.Second), current.Date)
}
