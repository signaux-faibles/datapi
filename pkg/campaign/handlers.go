package campaign

import (
	"datapi/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"regexp"
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

	pending, err := selectPending(c, CampaignID(id), []string{}, core.Page{10, 0})
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
	zone := zonesFromBoards(boards)
	campaigns, err := selectMatchingCampaigns(c, zone)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, campaignsp)
}

func matchConfigBoardSlugFilter(wekanDomainRegexp regexp.Regexp) func(libwekan.ConfigBoard) bool {
	return func(board libwekan.ConfigBoard) bool {
		return wekanDomainRegexp.MatchString(string(board.Slug()))
	}
}
