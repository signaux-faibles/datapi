package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"regexp"
	"time"
)

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
		&c.NBMyActions,
		&c.NBDone,
		&c.Zone,
		&c.BoardIDs,
	}
}

func selectMatchingCampaigns(ctx context.Context, zones map[string][]string,
	boardIDs map[string]string, username string) (campaigns Campaigns, err error) {
	campaigns = make(Campaigns, 0)
	err = db.Scan(ctx, &campaigns, sqlSelectMatchingCampaigns, zones, boardIDs, username)
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
	campaigns, err := selectMatchingCampaigns(c, zone, boardIDs, s.Username)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, campaigns)
}

func CampaignExists(ctx context.Context, campaignID CampaignID) bool {
	conn := db.Get()
	err := conn.QueryRow(ctx, "select from campaign where id=$1", campaignID).Scan()
	return err == nil
}

func GetCampaignWekanDomainRegexp(ctx context.Context, campaignID CampaignID) (wekanDomainRegexp string, err error) {
	conn := db.Get()
	err = conn.QueryRow(ctx, "select wekan_domain_regexp from campaign where id=$1", campaignID).Scan(&wekanDomainRegexp)
	if err != nil {
		return "", err
	}
	_, err = regexp.Compile(wekanDomainRegexp)
	return wekanDomainRegexp, err
}

func Create(ctx context.Context, title string, wekanDomainRegexp string, dateEnd time.Time) (CampaignID, error) {
	conn := db.Get()
	var id CampaignID
	err := conn.QueryRow(
		ctx,
		"insert into campaign (libelle, wekan_domain_regexp, date_end) values ($1, $2, $3) returning id",
		title,
		wekanDomainRegexp,
		dateEnd,
	).Scan(&id)

	return id, err
}
