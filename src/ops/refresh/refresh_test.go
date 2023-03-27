package refresh

import (
	"github.com/signaux-faibles/datapi/src/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var tuTime = time.Date(2023, 03, 10, 17, 41, 58, 651387237, time.Local)

func TestMain(m *testing.M) {
	test.FakeTime(tuTime)

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	// on peut placer ici du code de nettoyage si nécessaire
	test.UnfakeTime()

	os.Exit(code)
}

func Test_NewRefresh(t *testing.T) {
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
	test.FakeTime(expectedTime)

	current.run(expectedMessage)
	ass.Exactly(expectedMessage, current.Message)
	ass.Exactly(Running, current.Status)
	ass.Exactly(expectedTime, current.Date)
}
