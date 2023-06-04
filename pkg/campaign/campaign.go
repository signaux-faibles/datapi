package campaign

import (
	"context"
	"datapi/pkg/db"
	"datapi/pkg/utils"
	"fmt"
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
	NBTotal           int        `json:"nb_total"`
	NBPerimetre       int        `json:"nb_perimetre"`
	BoardIDs          []string   `json:"boardIDs"`
	Zone              []string   `json:"zone"`
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
		&c.NBTotal,
		&c.NBPerimetre,
		&c.Zone,
		&c.BoardIDs,
	}
}

func selectMatchingCampaigns(ctx context.Context, zones map[string][]string, boardIDs map[string]string) (Campaigns, error) {
	var allCampaigns = Campaigns{}
	fmt.Println(boardIDs)
	err := db.Scan(ctx, &allCampaigns, sqlSelectMatchingCampaigns, zones, boardIDs)
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

func idsFromBoards(boards []libwekan.ConfigBoard) map[string]string {
	ids := make(map[string]string)
	for _, board := range boards {
		ids[string(board.Board.Slug)] = string(board.Board.ID)
	}
	return ids
}
