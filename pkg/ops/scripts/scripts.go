// Package refresh : contient tout le code qui concerne l'exécution d'un `Refresh` Datapi,
// c'est-à-dire l'exécution du script sql configuré
package scripts

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/martian/log"
	"github.com/google/uuid"
)

type Script struct {
	Label string
	SQL   string
}

// pour les tests
var Wait5Seconds = Script{
	Label: "attends 5",
	SQL:   "SELECT pg_sleep(5);",
}

// pour les tests
var Fail = Script{
	Label: "sql invalide",
	SQL:   "sql invalide",
}

var list = sync.Map{}
var last = atomic.Value{}

// Run représente une exécution du script de `refresh` configuré dans l'application
type Run struct {
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
var Empty = Run{}

// NewRun crée un `Refresh`
func NewRun() *Run {
	r := Run{UUID: uuid.New()}
	r.save(Prepare, "🙏")
	last.Store(r.UUID)
	return &r
}

func NewScriptFrom(label, sql string) Script {
	return Script{
		Label: label,
		SQL:   sql,
	}
}

// Fetch : récupère un `Refresh`
func Fetch(id uuid.UUID) (Run, error) {
	value, found := list.Load(id)
	if found {
		return *value.(*Run), nil
	}
	return Run{}, errors.New("No refresh started with ID : " + id.String())
}

// FetchLast : récupère le dernier `Refresh`
func FetchLast() (Run, error) {
	val := last.Load()
	if val == nil {
		return Empty, nil
	}
	id := val.(uuid.UUID)
	return Fetch(id)
}

// FetchRefreshsWithState : récupère la liste des `Refresh` selon le `Status` passé en paramètre
func FetchRefreshsWithState(status Status) []Run {
	var retour []Run
	list.Range(func(k, v any) bool {
		other := *v.(*Run)
		if status == other.Status {
			retour = append(retour, other)
		}
		return true
	})
	return retour
}

func (r *Run) run(run string) {
	r.save(Running, run)
}

func (r *Run) fail(error string) {
	r.save(Failed, error)
}

func (r *Run) finish() {
	r.save(Finished, "👍")
}

func (r *Run) save(status Status, message string) {
	r.Status = status
	r.Message = message
	r.Date = time.Now()
	log.Infof("refresh script : %s", status)
	list.Store(r.UUID, r)
}

func (r *Run) String() string {
	return fmt.Sprintf(
		"Refresh{%s, date: %s, état: '%s', message: '%s'}",
		r.UUID.String(),
		r.Date.Format("2006-01-02 15:04:05.999999999"),
		r.Status,
		r.Message,
	)
}
