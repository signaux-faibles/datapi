package wekan

import (
	"github.com/signaux-faibles/libwekan"
	"time"
)

type GetCardForUserParams struct {
	Username  libwekan.Username `json:"-"`
	Type      string            `json:"type"`
	Zone      []string          `json:"zone"`
	BoardIDs  []string          `json:"boardIDs"`
	Labels    []string          `json:"labels"`
	Since     *time.Time        `json:"since"`
	Statuts   []string          `json:"statuts"`
	LabelMode bool              `json:"labelMode"`
}
