package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"time"
)

type CampaignID int

type Campaign struct {
	ID                CampaignID `json:"id"`
	Libelle           string     `json:"libelle"`
	DateEnd           time.Time  `json:"dateFin"`
	DateCreate        time.Time  `json:"dateCreate"`
	WekanDomainRegexp string     `json:"wekanDomainRegexp"`
	NBTotal           int        `json:"nbTotal"`
	NBPerimetre       int        `json:"nbPerimetre"`
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

func idsFromBoards(boards []libwekan.ConfigBoard) map[string]string {
	ids := make(map[string]string)
	for _, board := range boards {
		ids[string(board.Board.Slug)] = string(board.Board.ID)
	}
	return ids
}

func listCampaignsHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)
	boards := core.Kanban.SelectBoardsForUsername(libwekan.Username(s.Username))
	zone := zonesFromBoards(boards)
	boardIDs := idsFromBoards(boards)
	campaigns, err := selectMatchingCampaigns(c, zone, boardIDs)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, campaigns)
}
