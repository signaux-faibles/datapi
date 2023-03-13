package refresh

import (
	"github.com/signaux-faibles/datapi/src/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	test.FakeTime(test.TuTime)
	m.Run()
}

func Test_NewRefresh(t *testing.T) {
	ass := assert.New(t)
	current := New()
	ass.Equal(Prepare, current.Status)
	ass.NotNil(current.Message)
	ass.Equal(test.TuTime, current.Date)
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
	expectedTime := test.TuTime.Add(5 * time.Second)
	test.FakeTime(expectedTime)

	current.run(expectedMessage)
	ass.Exactly(expectedMessage, current.Message)
	ass.Exactly(Running, current.Status)
	ass.Exactly(test.TuTime.Add(5*time.Second), current.Date)
}
