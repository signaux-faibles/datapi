package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type CampaignEtablissement struct {
	ID                  int        `json:"id"`
	Siret               core.Siret `json:"siret"`
	Alert               string     `json:"alert"`
	Followed            bool       `json:"followed"`
	FirstAlert          bool       `json:"firstAlert"`
	EtatAdministratif   string     `json:"etatAdministratif"`
	Action              *string    `json:"action,omitempty"`
	Rank                int        `json:"rank"`
	RaisonSociale       string     `json:"raisonSociale"`
	RaisonSocialeGroupe *string    `json:"raisonSocialeGroupe,omitempty"`
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
	pending, err := selectPending(c, CampaignID(id), []string{}, core.Page{10, 0})
	if err != nil {
		c.JSON(http.StatusInternalServerError, "erreur inattendue: "+err.Error())
		return
	}
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
		&ce.Followed,
		&ce.FirstAlert,
		&ce.EtatAdministratif,
		&ce.Action,
		&ce.Rank,
	}
}

func selectPending(ctx context.Context, campaignID CampaignID, zone []string, page core.Page) (pending Pending, err error) {
	pending.Page = page.Number
	err = db.Scan(ctx, &pending, sqlSelectPendingEtablissement, page.Number, page.Size, campaignID, []string{"90", "75"})
	return pending, err
}
