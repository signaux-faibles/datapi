package campaign

import (
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"strconv"
)

func ConfigureEndpoint(path string, api *gin.Engine) {
	endpoint := api.Group(path, core.AuthMiddleware(), core.LogMiddleware)
	endpoint.GET("/list", listCampaignsHandler) // 1
	endpoint.GET("/pending/:campaignID", pendingHandler)
}

func pendingHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)

	id, err := strconv.Atoi(c.Param("campaignID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, "`"+c.Param("campaignID")+"` n'est pas un identifiant valide")
	}

	pending, err := selectPending(c, CampaignID(id), []string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, "erreur inattendue")
		return
	}
	if len(pending.Etablissements) == 0 {
		c.JSON(http.StatusNoContent, pending)
		return
	}
	c.JSON(http.StatusOK, pending)
}

func listCampaignsHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)

	boards := core.Kanban.SelectBoardsForUsername(libwekan.Username(s.Username))
	slugs := utils.Convert(boards, libwekan.ConfigBoard.Slug)
	campaigns, err := selectMatchingCampaigns(c, slugs, s.Roles)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	c.JSON(200, campaigns)
}

func zoneFromBoards(boards []libwekan.ConfigBoard) []string {
	var zone []string
	for _, board := range boards {
		for _, swimlane := range board.Swimlanes {
			departement := swimlane.Title
		}
	}
	return nil
}
