package campaign

import (
	"datapi/pkg/core"
	"github.com/signaux-faibles/libwekan"
	"time"
)

type CampaignEtablissementActionID int

type Action struct {
	action string
	detail string
}

type CampaignID int

type Campaign struct {
	ID                CampaignID `json:"id"`
	Libelle           string     `json:"libelle"`
	DateEnd           time.Time  `json:"dateFin"`
	DateCreate        time.Time  `json:"dateCreate"`
	WekanDomainRegexp string     `json:"wekanDomainRegexp"`
	NBPerimetre       int        `json:"nbPerimetre"`
	NBPending         int        `json:"nbPending"`
	NBTake            int        `json:"nbTake"`
	NBMyActions       int        `json:"nbMyActions"`
	NBDone            int        `json:"nbDone"`
	BoardIDs          []string   `json:"boardIDs"`
	Zone              []string   `json:"zone"`
}

type Campaigns []*Campaign

type Message struct {
	CampaignID              CampaignID               `json:"campaignID"`
	CampaignEtablissementID *CampaignEtablissementID `json:"campaignEtablissementID"`
	Zone                    []string                 `json:"zone"`
	Type                    string                   `json:"type"`
	Username                string                   `json:"username"`
}

type MyActions struct {
	Etablissements    []*CampaignEtablissement `json:"etablissements"`
	NbTotal           int                      `json:"nbTotal"`
	Page              int                      `json:"page"`
	PageMax           int                      `json:"pageMax"`
	PageSize          int                      `json:"pageSize"`
	WekanDomainRegexp string                   `json:"-"`
}

type IDs struct {
	CampaignID
	CampaignEtablissementID
}

type CampaignEtablissementID int

type CampaignEtablissement struct {
	ID                  CampaignEtablissementID `json:"id"`
	CampaignID          int                     `json:"campaignID"`
	Siret               core.Siret              `json:"siret"`
	Alert               *string                 `json:"alert"`
	Followed            bool                    `json:"followed"`
	FirstAlert          bool                    `json:"firstAlert"`
	EtatAdministratif   string                  `json:"etatAdministratif"`
	CodeDepartement     string                  `json:"codeDepartement"`
	Action              *string                 `json:"action,omitempty"`
	Rank                int                     `json:"rank"`
	RaisonSociale       string                  `json:"raisonSociale"`
	RaisonSocialeGroupe *string                 `json:"raisonSocialeGroupe,omitempty"`
	Username            *string                 `json:"username,omitempty"`
	CardID              *libwekan.CardID        `json:"cardID,omitempty"`
	Description         *string                 `json:"description,omitempty"`
	Detail              *string                 `json:"detail,omitempty"`
}

type Pending struct {
	Etablissements    []*CampaignEtablissement `json:"etablissements"`
	NbTotal           int                      `json:"nbTotal"`
	Page              int                      `json:"page,omitempty"`
	PageMax           int                      `json:"pageMax,omitempty"`
	PageSize          int                      `json:"pageSize,omitempty"`
	WekanDomainRegexp string                   `json:"-"`
}

type TakenActions struct {
	Etablissements    []*CampaignEtablissement `json:"etablissements"`
	NbTotal           int                      `json:"nbTotal"`
	WekanDomainRegexp string                   `json:"-"`
}

type actionParam struct {
	Detail string `json:"detail"`
}

type Event struct {
	Message       chan Message
	NewClients    chan chan Message
	ClosedClients chan chan Message
	TotalClients  map[chan Message]bool
}

type ClientChan chan Message

type BoardZones map[string][]string
type Zone []string
