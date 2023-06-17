package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"datapi/pkg/kanban"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"strconv"
)

type CampaignEtablissementID int

type CampaignEtablissement struct {
	ID                  CampaignEtablissementID `json:"id"`
	CampaignID          int                     `json:"campaignID"`
	Siret               core.Siret              `json:"siret"`
	Alert               string                  `json:"alert"`
	Followed            bool                    `json:"followed"`
	FirstAlert          bool                    `json:"firstAlert"`
	EtatAdministratif   string                  `json:"etatAdministratif"`
	Action              *string                 `json:"action,omitempty"`
	Rank                int                     `json:"rank"`
	RaisonSociale       string                  `json:"raisonSociale"`
	RaisonSocialeGroupe *string                 `json:"raisonSocialeGroupe,omitempty"`
	Username            *string                 `json:"username,omitempty"`
}

type Pending struct {
	Etablissements []*CampaignEtablissement `json:"etablissements"`
	NbTotal        int                      `json:"nbTotal"`
	Page           int                      `json:"page"`
	PageMax        int                      `json:"pageMax"`
	PageSize       int                      `json:"pageSize"`
}

func pendingHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)
	id, err := strconv.Atoi(c.Param("campaignID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, "`"+c.Param("campaignID")+"` n'est pas un identifiant valide")
		return
	}
	username := libwekan.Username(s.Username)
	boards := kanban.SelectBoardsForUser(username)
	zones := zonesFromBoards(boards)
	pending, err := selectPending(c, CampaignID(id), zones, core.Page{10, 0}, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "erreur inattendue: "+err.Error())
		return
	}
	fmt.Println(pending)
	if len(pending.Etablissements) == 0 {
		c.JSON(http.StatusNoContent, pending)
		return
	}
	c.JSON(http.StatusOK, pending)
}

func (p *Pending) Tuple() []interface{} {
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
	}
}

func selectPending(ctx context.Context, campaignID CampaignID, zone BoardZones, page core.Page, username libwekan.Username) (pending Pending, err error) {
	err = db.Scan(ctx, &pending, sqlSelectPendingEtablissement, campaignID, zone, username)
	return pending, err
}
