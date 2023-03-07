package refresh

import (
	"github.com/google/martian/log"
	"github.com/google/uuid"
	"sync"
	"time"
)

var refreshingState = sync.Map{}

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

func New(id uuid.UUID) *Refresh {
	r := new(Refresh)
	r.Id = id
	r.save(Prepare)
	return r
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
	refreshingState.Store("refreshing", r)
}
