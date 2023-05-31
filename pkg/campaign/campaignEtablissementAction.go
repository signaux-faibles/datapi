package campaign

import (
	"github.com/signaux-faibles/libwekan"
	"time"
)

type CampaignEtablissementActionID int

type CampaignEtablissementAction struct {
	ID                      CampaignEtablissementActionID `json:"id"`
	CampaignEtablissementID `json:"campaignEtablissementID"`
	Username                libwekan.Username
	Action                  string
	DateAction              time.Time
	Metadata                interface{}
}
