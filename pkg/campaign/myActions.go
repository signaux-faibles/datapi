package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"strconv"
)

type MyActions struct {
	Etablissements []*CampaignEtablissement `json:"etablissements"`
	NbTotal        int                      `json:"nbTotal"`
	Page           int                      `json:"page"`
	PageMax        int                      `json:"pageMax"`
	PageSize       int                      `json:"pageSize"`
}

func (p *MyActions) Tuple() []interface{} {
	var ce CampaignEtablissement
	p.Etablissements = append(p.Etablissements, &ce)
	return []interface{}{
		&p.NbTotal,
		&ce.Siret,
		&ce.RaisonSociale,
		&ce.RaisonSocialeGroupe,
		&ce.Alert,
		&ce.ID,
		&ce.CampaignID,
		&ce.Followed,
		&ce.FirstAlert,
		&ce.EtatAdministratif,
		&ce.Action,
		&ce.Rank,
		&ce.CodeDepartement,
	}
}

func myActionsHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)

	campaignID, err := strconv.Atoi(c.Param("campaignID"))
	if err != nil {
		c.JSON(400, `/campaign/myactions/:campaignID: le parametre campaignID doit être un entier`)
		return
	}
	zone := zoneForUser(libwekan.Username(s.Username))
	page := core.Page{10, 0}
	myActions, err := selectMyActions(c, CampaignID(campaignID), zone, page, libwekan.Username(s.Username))
	c.JSON(200, myActions)
}

func selectMyActions(ctx context.Context, campaignID CampaignID, zone []string, page core.Page, username libwekan.Username) (myActions MyActions, err error) {
	err = db.Scan(ctx, &myActions, sqlSelectMyActions, campaignID, zone, username)
	return myActions, err
}
