package campaign

import (
	"datapi/pkg/core"
)

type CampaignEtablissementID int

type CampaignEtablissementMetadata interface{}

type CampaignEtablissement struct {
	ID         CampaignEtablissementID       `json:"id"`
	CampaignID CampaignID                    `json:"campaignID"`
	Siret      core.Siret                    `json:"siret"`
	Metadata   CampaignEtablissementMetadata `json:"metadata"`
	Actions    []CampaignEtablissementAction `json:"actions"`
}
