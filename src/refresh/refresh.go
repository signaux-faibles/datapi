// Package refresh : contient tout le code qui concerne l'exécution d'un `Refresh` Datapi,
// c'est-à-dire l'exécution du script sql configuré
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

// Refresh : représente une exécution du script de `refresh` configuré dans l'application
type Refresh struct {
	UUID    uuid.UUID `json:"UUID,omitempty"`
	Status  Status    `json:"Status,omitempty"`
	Date    time.Time `json:"Date,omitempty"`
	Message string    `json:"Message,omitempty"`
}

// Status : propriété du `Refresh` qui représente son état
type Status string

const (
	// Prepare : état du `Refresh` en préparation
	Prepare Status = "prepare"
	// Running : état du `Refresh` lorsque le SQL est en cours d'exécution
	Running Status = "running"
	// Failed : état du `Refresh` lorsqu'une erreur est levée pendant son exécution'
	Failed Status = "failed"
	// Finished : état du `Refresh` lorsque tout s'est bien passé
	Finished Status = "finished"
)

// Empty : représente un `Refresh` nul
var Empty = Refresh{}

// New : crée un `Refresh`
func New(uuid uuid.UUID) *Refresh {
	r := Refresh{UUID: uuid}
	r.save(Prepare, "🙏")
	last.Store(r)
	return &r
}

// Fetch : récupère un `Refresh`
func Fetch(id uuid.UUID) (Refresh, error) {
	value, found := states.Load(id)
	if found {
		return *value.(*Refresh), nil
	}
	return Refresh{}, errors.New("No refresh started with ID : " + id.String())
}

// FetchLast : récupère le dernier `Refresh`
func FetchLast() Refresh {
	val := last.Load()
	if val == nil {
		return Empty
	}
	return val.(Refresh)
}

// FetchRefreshsWithState : récupère la liste des `Refresh` selon le `Status` passé en paramètre
func FetchRefreshsWithState(status Status) []Refresh {
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
	states.Store(r.UUID, r)
}

func (r Refresh) String() string {
	return fmt.Sprintf(
		"Refresh{%s, date: %s, état: '%s', message: '%s'}",
		r.UUID,
		r.Date.Format("2006-01-02 15:04:05.999999999"),
		r.Status,
		r.Message,
	)
}
