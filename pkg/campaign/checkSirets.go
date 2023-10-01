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

type CheckedSirets struct {
	Sirets []*CheckedSiret `json:"sirets"`
}

type CheckedSiret struct {
	Siret           *core.Siret `json:"siret"`
	RaisonSociale   *string     `json:"raisonSociale"`
	CodeDepartement *string     `json:"codeDepartement"`
	Status          *string     `json:"status"`
}

func (checkedSirets *CheckedSirets) Tuple() []interface{} {
	var checkedSiret CheckedSiret
	checkedSirets.Sirets = append(checkedSirets.Sirets, &checkedSiret)
	return []interface{}{
		&checkedSiret.Siret,
		&checkedSiret.Status,
		&checkedSiret.RaisonSociale,
		&checkedSiret.CodeDepartement,
	}
}

type CheckSiretsParams struct {
	Sirets []core.Siret `json:"sirets"`
}

func checkSiretsHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)
	campaignIDParam := c.Param("campaignID")
	campaignID, err := strconv.Atoi(campaignIDParam)
	var params CheckSiretsParams
	c.Bind(&params)
	result, err := checkSirets(c, CampaignID(campaignID), params.Sirets, s.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func checkSirets(ctx context.Context, campaignID CampaignID, sirets []core.Siret, username string) (CheckedSirets, error) {
	var checkedSiret CheckedSirets
	checkedSiret.Sirets = make([]*CheckedSiret, 0)
	boards := core.Kanban.SelectBoardsForUsername(libwekan.Username(username))
	zones := zonesFromBoards(boards)
	err := db.Scan(ctx, &checkedSiret, sqlCheckSirets, campaignID, sirets, zones)
	return checkedSiret, err
}
