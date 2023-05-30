package campaign

import (
	"datapi/pkg/core"
)

type CampaignEtablissementID int

type CampaignEtablissement struct {
	ID      CampaignEtablissementID       `json:"id"`
	Siret   core.Siret                    `json:"siret"`
	Actions []CampaignEtablissementAction `json:"actions"`
}

type CampaignDetails []CampaignEtablissement

func (c CampaignDetails) NewItem() []interface{} {
	return nil
}

type CampaignEtablissementAction struct {
}

//
//func selectCampaignDetails(ctx context.Context, campaignID int, db *pgxpool.Pool) (CampaignDetails, error) {
//	rows, err := db.Query(ctx, `select * from campaign c
//inner join campaign_etablissement ce on ce.id_campaign = c.id
//left join campaign_etablissement_action cea on cea.id_campaign_etablissement = ce.id`)
//}
