package refresh

import (
	"github.com/google/uuid"
	"sync"
	"time"
)

var refreshingState = sync.Map{}

type Refresh struct {
	Id     uuid.UUID
	Status Status
	Date   time.Time
}

type Status string

const (
	Prepare  Status = "prepare"
	Running  Status = "is running"
	Failed   Status = "has failed"
	Finished Status = "has finished"
)

func New() *Refresh {
	r := new(Refresh)
	r.Id = uuid.New()
	r.save(Prepare)
	return r
}

func (r Refresh) Run() {
	r.save(Running)
}

func (r Refresh) Fail() {
	r.save(Failed)
}

func (r Refresh) Finish() {
	r.save(Finished)
}

func (r Refresh) save(status Status) {
	r.Status = status
	r.Date = time.Now()
	refreshingState.Store("refreshing", r)
}
