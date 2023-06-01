package campaign

import (
	"github.com/lib/pq/hstore"
	"github.com/signaux-faibles/libwekan"
	"time"
)

type CampaignEtablissementActionID int

type CampaignEtablissementAction struct {
	ID                      *CampaignEtablissementActionID `json:"id"`
	CampaignEtablissementID *CampaignEtablissementID       `json:"campaignEtablissementID"`
	Username                *libwekan.Username
	Action                  *string
	DateAction              *time.Time
	Metadata                *hstore.Hstore
}
