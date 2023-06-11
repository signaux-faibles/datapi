package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"errors"
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
	zone := zoneForUser(username)
	pending, err := selectPending(c, CampaignID(id), zone, core.Page{10, 0}, username)
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
		&ce.CampaignID,
		&ce.Followed,
		&ce.FirstAlert,
		&ce.EtatAdministratif,
		&ce.Action,
		&ce.Rank,
	}
}

func selectPending(ctx context.Context, campaignID CampaignID, zone []string, page core.Page, username libwekan.Username) (pending Pending, err error) {
	err = db.Scan(ctx, &pending, sqlSelectPendingEtablissement, campaignID, zone, username)
	return pending, err
}

func takePendingHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)

	campaignID, err := strconv.Atoi(c.Param("campaignID"))
	if err != nil {
		c.JSON(400, `/campaign/take/:campaignID/:campaignEtablissementID: le parametre campaignID doit être un entier`)
		return
	}
	campaignEtablissementID, err := strconv.Atoi(c.Param("campaignEtablissementID"))
	if err != nil {
		c.JSON(400, `/campaign/take/:campaignID/:campaignEtablissementID: le parametre campaignEtablissementID doit être un entier`)
		return
	}

	zone := zoneForUser(libwekan.Username(s.Username))

	err = takePending(
		c,
		CampaignID(campaignID),
		CampaignEtablissementID(campaignEtablissementID),
		libwekan.Username(s.Username),
		zone,
	)

	if errors.As(err, &PendingNotFoundError{}) {
		c.JSON(http.StatusUnprocessableEntity, "établissement indisponible pour cette action")
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, "erreur imprévue")
	} else {
		c.JSON(http.StatusOK, "ok")
	}
}

func takePending(ctx context.Context, campaignID CampaignID, campaignEtablissementID CampaignEtablissementID, username libwekan.Username, zone Zone) error {
	var campaignEtablissementActionID *int
	dbConn := db.Get()
	row := dbConn.QueryRow(ctx, sqlTakePendingEtablissement, campaignID, campaignEtablissementID, username, zone)
	err := row.Scan(&campaignEtablissementActionID)
	if campaignEtablissementActionID == nil {
		return PendingNotFoundError{err: errors.New("aucun id retourné par l'insert")}
	} else if err != nil {
		return err
	}
	return nil
}
