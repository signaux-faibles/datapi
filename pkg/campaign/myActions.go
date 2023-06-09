package campaign

import (
	"datapi/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
)

type MyActions struct {
	Etablissements []*CampaignEtablissement `json:"etablissements"`
	NbTotal        int                      `json:"nbTotal"`
	Page           int                      `json:"page"`
	PageMax        int                      `json:"pageMax"`
	PageSize       int                      `json:"pageSize"`
}

func myActionsHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)

	var myCampaignEtablissementActions MyActions
	c.JSON(200, myCampaignEtablissementActions)
}

func selectMyActions(ctx, campaignID CampaignID, username libwekan.Username) (MyActions, error) {
	return MyActions{}, nil
}
