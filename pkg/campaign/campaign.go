package campaign

import (
	"context"
	"datapi/pkg/db"
	"datapi/pkg/utils"
	"github.com/signaux-faibles/libwekan"
	"strings"
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

func selectMatchingCampaigns(ctx context.Context, zones map[string][]string) (Campaigns, error) {
	var allCampaigns = Campaigns{}
	err := db.Scan(ctx, &allCampaigns, sqlSelectMatchingCampaigns, zones)
	return allCampaigns, err
}

func zonesFromBoards(boards []libwekan.ConfigBoard) map[string][]string {
	zones := make(map[string][]string)
	for _, board := range boards {
		swimlanes := utils.GetValues(board.Swimlanes)
		zone := utils.Convert(swimlanes, func(s libwekan.Swimlane) string {
			return strings.Split(string(s.Title), " (")[0]
		})
		zones[string(board.Board.Slug)] = zone
	}
	return zones
}
