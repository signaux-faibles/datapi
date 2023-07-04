package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"strconv"
)

type AllActions struct {
	Etablissements []*CampaignEtablissement `json:"etablissements"`
	NbTotal        int                      `json:"nbTotal"`
	Page           int                      `json:"page"`
	PageMax        int                      `json:"pageMax"`
	PageSize       int                      `json:"pageSize"`
}

func (p *AllActions) Tuple() []interface{} {
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
		&ce.Username,
	}
}

func takenActionsHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)
	campaignID, err := strconv.Atoi(c.Param("campaignID"))
	if err != nil {
		c.JSON(400, `/campaign/actions/taken/:campaignID: le parametre campaignID doit être un entier`)
		return
	}
	zone := zoneForUser(libwekan.Username(s.Username))
	allActions, err := selectTakenActions(c, CampaignID(campaignID), zone, libwekan.Username(s.Username))

	if err != nil {
		c.JSON(500, err.Error())
	} else {
		c.JSON(200, allActions)
	}
}

func selectTakenActions(ctx context.Context, campaignID CampaignID, zone []string, username libwekan.Username) (allActions AllActions, err error) {
	err = db.Scan(ctx, &allActions, sqlSelectTakenActions, campaignID, zone, username)
	return allActions, err
}
