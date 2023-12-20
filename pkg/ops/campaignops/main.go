package campaignops

import (
	"context"
	"datapi/pkg/campaign"
	"datapi/pkg/core"
	"datapi/pkg/db"
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"time"
)

// ConfigureEndpoint configure l'endpoint du package `ops`
func ConfigureEndpoint(kanbanService core.KanbanService) func(endpoint *gin.RouterGroup) {
	return func(endpoint *gin.RouterGroup) {
		endpoint.POST("/new", newCampaignHandler(kanbanService))
	}
}

type NewCampaignParams struct {
	DateFin            time.Time           `json:"dateFin"`
	FromCampaignID     campaign.CampaignID `json:"fromCampaignID"`
	FromListeDetection string              `json:"fromListeDetection"`
}

func newCampaignHandler(kanbanService core.KanbanService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var params NewCampaignParams

		err := c.Bind(&params)
		fmt.Println(params)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
		}

		err = checkParams(c, params)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
		}

		wekanDomainRegexp, err := campaign.GetCampaignWekanDomainRegexp(c, params.FromCampaignID)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
		}

		siretsFromListe, err := selectSiretsFromListe(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		slog.Info("sirets provenant de la liste", slog.Int("nb_sirets", len(siretsFromListe)))

		siretsFromReports, err := selectSiretsFromAction(c, params, "withdraw")
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		slog.Info("sirets provenant des reports", slog.Int("siretsFromReport", len(siretsFromReports)))

		siretsFromEncours, err := selectSiretsFromAction(c, params, "take")
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		slog.Info("sirets provenant des en cours", slog.Int("siretsFromEncours", len(siretsFromEncours)))

		siretsFromAccompagnement, err := selectSiretsFromAccompagnement(c, wekanDomainRegexp, kanbanService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		slog.Info("sirets provenant des accompagnements wekan", slog.Int("siretsFromAccompagnement", len(siretsFromAccompagnement)))

		// Récupération des données de contexte (
		//    - siretsFromListe à insérer
		//       - depuis la liste
		//       - depuis les reports (3 mois/6 mois/9 mois/definitif)
		//    - siretsFromListe à exclure
		//       - entreprises accompagnées via wekan

		// Calcul des données à insérer

		// Insertion des données
	}
}

func selectSiretsFromAccompagnement(ctx context.Context, wekanDomainRegexp string, kanbanService core.KanbanService) ([]core.Siret, error) {
	return kanbanService.SelectSiretsFromListeAndDomainRegexp(ctx, wekanDomainRegexp, "Accompagnement en cours")
}

//go:embed sql/siretsFromAction.sql
var sqlSiretsFromAction string

func selectSiretsFromAction(ctx context.Context, params NewCampaignParams, actionName string) ([]campaignAction, error) {
	conn := db.Get()
	rows, err := conn.Query(ctx, sqlSiretsFromAction, params.FromCampaignID, actionName)
	if err != nil {
		return nil, err
	}
	var actions []campaignAction
	for rows.Next() {
		var action campaignAction
		err = rows.Scan(&action.siret, &action.username, &action.actionDetail)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, nil
}

type campaignAction struct {
	siret        core.Siret
	actionDetail string
	username     string
}

//go:embed sql/siretsFromListe.sql
var sqlSiretsFromListe string

func selectSiretsFromListe(ctx context.Context, params NewCampaignParams) ([]core.Siret, error) {
	conn := db.Get()
	rows, err := conn.Query(ctx, sqlSiretsFromListe, params.FromListeDetection)
	if err != nil {
		return nil, err
	}
	var sirets []core.Siret
	for rows.Next() {
		var siret core.Siret
		err = rows.Scan(&siret)
		if err != nil {
			return nil, err
		}
		sirets = append(sirets, siret)
	}
	return sirets, nil
}

func checkParams(ctx context.Context, params NewCampaignParams) error {
	if params.DateFin.Before(time.Now()) {
		return fmt.Errorf("dateFin doit être dans le futur : %s", params.DateFin.Format(time.DateOnly))
	}
	if !campaign.CampaignExists(ctx, params.FromCampaignID) {
		return fmt.Errorf("l'id de campagne n'existe pas : %d", params.FromCampaignID)
	}
	if !core.ListeExists(ctx, params.FromListeDetection) {
		return fmt.Errorf("la liste n'existe pas: %s", params.FromListeDetection)
	}
	return nil
}
