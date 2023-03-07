package refresh

import (
	"errors"
	"fmt"
	"github.com/google/martian/log"
	"github.com/google/uuid"
	"sync"
	"sync/atomic"
	"time"
)

var states = sync.Map{}
var last = atomic.Value{}

type Refresh struct {
	Uuid    uuid.UUID
	Status  Status
	Date    time.Time
	Message string
}

type Status string

const (
	Prepare  Status = "prepare"
	Running  Status = "is running"
	Failed   Status = "has failed"
	Finished Status = "has finished"
)

var Empty = Refresh{}

func new(uuid uuid.UUID) *Refresh {
	r := Refresh{Uuid: uuid}
	r.save(Prepare, "🙏")
	last.Store(r)
	return &r
}

func Fetch(id uuid.UUID) (Refresh, error) {
	value, found := states.Load(id)
	if found {
		return *value.(*Refresh), nil
	}
	return Refresh{}, errors.New("No refreshing with ID : " + id.String())
}

func FetchLastRefreshState() Refresh {
	val := last.Load()
	if val == nil {
		return Empty
	}
	return val.(Refresh)
}

func FetchRefreshWithState(status Status) []Refresh {
	var retour []Refresh
	states.Range(func(k, v any) bool {
		other := *v.(*Refresh)
		if status == other.Status {
			retour = append(retour, other)
		}
		return true
	})
	return retour
}

func (r *Refresh) run(run string) {
	r.save(Running, run)
}

func (r *Refresh) fail(error string) {
	r.save(Failed, error)
}

func (r *Refresh) finish() {
	r.save(Finished, "👍")
}

func (r *Refresh) save(status Status, message string) {
	r.Status = status
	r.Message = message
	r.Date = time.Now()
	log.Infof("refresh script : %s", status)
	states.Store(r.Uuid, r)
}

func (r Refresh) String() string {
	return fmt.Sprintf(
		"Refresh{%s, date: %s, état: '%s', message: '%s'}",
		r.Uuid,
		r.Date.Format("2006-01-02 15:04:05.999999999"),
		r.Status,
		r.Message,
	)
}
