package campaign

import (
	"context"
	"datapi/pkg/db"
)

type CampaignDetail struct {
	Campaign
	CampaignEtablissement
	CampaignEtablissementAction
}
type CampaignDetails []*CampaignDetail

//id_campaign, c.libelle, c.wekan_domain_regexp, c.date_end, c.date_create,
//ce.id as id_campaign_etablissement, ce.siret, ce.metadata,
//cea.id as id_campaign_etablissement_action, cea.username, cea.date_action, cea.action, cea.metadata

func (cds *CampaignDetails) Tuple() []interface{} {
	c := Campaign{}
	ce := CampaignEtablissement{}
	cea := CampaignEtablissementAction{}
	cd := CampaignDetail{
		Campaign:                    c,
		CampaignEtablissement:       ce,
		CampaignEtablissementAction: cea,
	}

	*cds = append(*cds, &cd)
	return []interface{}{
		&c.ID,
		&c.Libelle,
		&c.WekanDomainRegexp,
		&c.DateEnd,
		&c.DateCreate,
		&c.BoardSlugs,
		&ce.ID,
		&ce.CampaignID,
		&ce.Siret,
		&ce.Metadata,
		&cea.ID,
		&cea.CampaignEtablissementID,
		&cea.Username,
		&cea.DateAction,
		&cea.Action,
		&cea.Metadata,
	}
}

func selectCampaignDetailsWithCampaignIDAndZone(ctx context.Context, id CampaignID, zone []string) (CampaignDetails, error) {
	sql := `select c.id as id_campaign, c.libelle, c.wekan_domain_regexp, c.date_end, c.date_create,
      ce.id as id_campaign_etablissement, c.id, ce.siret, ce.metadata,
      cea.id as id_campaign_etablissement_action, ce.id, cea.username, cea.date_action, cea.action, cea.metadata
    from campaign c
    inner join campaign_etablissement ce on c.id = ce.id_campaign
    inner join v_summaries s on s.siret = ce.siret
    left join campaign_etablissement_action cea on ce.id = cea.id_campaign_etablissement
    where c.id = $1 and s.code_departement = any($2);`

	var campaignDetails CampaignDetails
	err := db.Scan(ctx, &campaignDetails, sql, id, zone)

	return campaignDetails, err
}
