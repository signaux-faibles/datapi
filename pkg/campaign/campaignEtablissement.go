package campaign

import (
	"datapi/pkg/core"
	"github.com/lib/pq/hstore"
)

type CampaignEtablissementID int

type CampaignEtablissementMetadata interface{}

type CampaignEtablissement struct {
	ID         CampaignEtablissementID       `json:"id"`
	CampaignID CampaignID                    `json:"campaignID"`
	Siret      core.Siret                    `json:"siret"`
	Metadata   hstore.Hstore                 `json:"metadata"`
	Actions    []CampaignEtablissementAction `json:"actions"`
}
