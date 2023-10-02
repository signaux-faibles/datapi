package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"strconv"
)

type AddedSirets struct {
	Sirets []*AddedSiret `json:"sirets"`
}

type AddedSiret struct {
	Siret           core.Siret `json:"siret"`
	Status          string     `json:"status"`
	RaisonSociale   string     `json:"raisonSociale"`
	CodeDepartement string     `json:"codeDepartement"`
	Outcome         string     `json:"outcome"`
}

func (addedSirets AddedSirets) Tuple() []interface{} {
	var addedSiret AddedSiret
	addedSirets.Sirets = append(addedSirets.Sirets, &addedSiret)
	return []interface{}{
		&addedSiret.Siret,
		&addedSiret.Status,
		&addedSiret.RaisonSociale,
		&addedSiret.CodeDepartement,
		&addedSiret.Outcome,
	}
}

func addSiretsHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)
	campaignIDParam := c.Param("campaignID")
	campaignID, err := strconv.Atoi(campaignIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, "campaignID doit être un nombre entier")
		return
	}
	var params CheckSiretsParams
	err = c.Bind(&params)
	if err != nil {
		return
	}
	message, err := addSirets(c, CampaignID(campaignID), params.Sirets, s.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	stream.Message <- message
	c.JSON(http.StatusOK, message)
}

func addSirets(ctx context.Context, campaignID CampaignID, sirets []core.Siret, username string) (Message, error) {
	var addedSirets AddedSirets
	addedSirets.Sirets = make([]*AddedSiret, 0)
	boards := core.Kanban.SelectBoardsForUsername(libwekan.Username(username))
	zones := zonesFromBoards(boards)
	err := db.Scan(ctx, &addedSirets, sqlAddSirets, campaignID, sirets, zones, username)
	message := Message{
		CampaignID: campaignID,
		Zone:       zoneFromBoardZones(zones),
		Type:       "addSiret",
		Username:   username,
	}
	return message, err
}
