package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
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
	NBPerimetre       int        `json:"nbPerimetre"`
	NBPending         int        `json:"nbPending"`
	NBTake            int        `json:"nbTake"`
	NBDone            int        `json:"nbDone"`
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
		&c.NBPerimetre,
		&c.NBPending,
		&c.NBTake,
		&c.NBDone,
		&c.Zone,
		&c.BoardIDs,
	}
}

func selectMatchingCampaigns(ctx context.Context, zones map[string][]string, boardIDs map[string]string) (campaigns Campaigns, err error) {
	campaigns = make(Campaigns, 0)
	err = db.Scan(ctx, &campaigns, sqlSelectMatchingCampaigns, zones, boardIDs)
	return campaigns, err
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
