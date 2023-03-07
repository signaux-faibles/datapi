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
	Id      uuid.UUID
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

func New(id uuid.UUID) *Refresh {
	r := new(Refresh)
	r.Id = id
	r.save(Prepare)
	last.Store(r)
	return r
}

func Fetch(id uuid.UUID) (Refresh, error) {
	value, found := states.Load(id)
	if found {
		return value.(Refresh), nil
	}
	return Refresh{}, errors.New("No refreshing with ID : " + id.String())
}

func FetchLast() Refresh {
	val := last.Load()
	if val == nil {
		return Empty
	}
	return *val.(*Refresh)
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

func (r *Refresh) Run(run string) {
	r.Message = run
	r.save(Running)
}

func (r *Refresh) Fail(error string) {
	r.Message = error
	r.save(Failed)
}

func (r *Refresh) Finish() {
	r.save(Finished)
}

func (r *Refresh) save(status Status) {
	r.Status = status
	r.Date = time.Now()
	log.Infof("refresh script : %s", status)
	states.Store(r.Id, r)
}

func (r Refresh) String() string {
	return fmt.Sprintf(
		"Refresh{%s, date: %s, état: '%s', message: '%s'}",
		r.Id,
		r.Date.Format("2006-01-02 15:04:05 -070000"),
		r.Status,
		r.Message,
	)
}
