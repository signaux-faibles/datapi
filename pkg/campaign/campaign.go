package campaign

import (
	"context"
	"datapi/pkg/db"
	"github.com/signaux-faibles/libwekan"
	"time"
)

type CampaignID int

type Campaign struct {
	ID                CampaignID `json:"id"`
	Libelle           string     `json:"libelle"`
	DateEnd           time.Time  `json:"date_fin"`
	DateCreate        time.Time  `json:"date_create"`
	WekanDomainRegexp string     `json:"wekan_domain_regexp"`
	BoardSlugs        []string   `json:"slugs"`
	NBTotal           int        `json:"nb_total"`
	NBPerimetre       int        `json:"nb_perimetre"`
}

type Campaigns []*Campaign

func (cs *Campaigns) Tuple() []interface{} {
	c := Campaign{}
	*cs = append(*cs, &c)
	return []interface{}{
		&c.ID,
		&c.Libelle,
		&c.WekanDomainRegexp,
		&c.DateEnd,
		&c.DateCreate,
		&c.BoardSlugs,
		&c.NBTotal,
		&c.NBPerimetre,
	}
}

func selectMatchingCampaigns(ctx context.Context, boardSlugs []libwekan.BoardSlug, roles []string) (Campaigns, error) {
	var allCampaigns = Campaigns{}
	err := db.Scan(ctx, &allCampaigns, sqlSelectMatchingCampaigns, boardSlugs, roles)
	return allCampaigns, err
}
