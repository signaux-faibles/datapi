package campaignops

import (
	"context"
	"datapi/pkg/campaign"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"datapi/pkg/utils"
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/signaux-faibles/libwekan"
	"log/slog"
	"net/http"
	"slices"
	"time"
)

// ConfigureEndpoint configure l'endpoint du package `ops`
func ConfigureEndpoint(kanbanService core.KanbanService) func(endpoint *gin.RouterGroup) {
	return func(endpoint *gin.RouterGroup) {
		endpoint.POST("/new", newCampaignHandler(kanbanService))
	}
}

type newCampaignParams struct {
	DateFin            time.Time           `json:"dateFin"`
	FromCampaignID     campaign.CampaignID `json:"fromCampaignID"`
	FromListeDetection string              `json:"fromListeDetection"`
}

type newCampaignSirets struct {
	config                         libwekan.Config
	fromWekanAccompagnementEnCours []libwekan.Card
	fromWekanAnalyseEnCours        []libwekan.Card
	fromListe                      []core.Siret
	fromCampaignReports            []campaignAction
	fromCampaignEncours            []campaignAction
}

type campaignAction struct {
	siret        core.Siret
	actionDetail string
	username     string
}

func (a campaignAction) getSiret() core.Siret {
	return a.siret
}

func deleteSiretsFromSiretsSlice(sirets []core.Siret) func(core.Siret) bool {
	return func(siret core.Siret) bool {
		for _, current := range sirets {
			if current == siret {
				return true
			}
		}
		return false
	}
}

func deleteCampaignActionFromSiretsSlice(sirets []core.Siret) func(campaignAction) bool {
	return func(action campaignAction) bool {
		for _, current := range sirets {
			if current == action.siret {
				return true
			}
		}
		return false
	}
}

func deleteCampaignReportDelai3mois(action campaignAction) bool {
	return action.actionDetail == "delai_3mois"
}

func updateCampaignReport(action campaignAction) campaignAction {
	switch action.actionDetail {
	case "delai_9mois":
		action.actionDetail = "delai_6mois"
	case "delai_6mois":
		action.actionDetail = "delai_3mois"
	}
	return action
}

func siretWithNameFunc(config libwekan.Config) func(card libwekan.Card) core.Siret {
	return func(card libwekan.Card) core.Siret {
		return core.Siret(config.CustomFieldWithName(card, "SIRET"))
	}
}

func (n newCampaignSirets) fromWekanAccompagnementEnCoursSirets() []core.Siret {
	return utils.Convert(n.fromWekanAccompagnementEnCours, siretWithNameFunc(n.config))
}

func (n newCampaignSirets) fromWekanAnalyseEnCoursSirets() []core.Siret {
	return utils.Convert(n.fromWekanAnalyseEnCours, siretWithNameFunc(n.config))
}

func (n newCampaignSirets) siretsToInsert() []core.Siret {
	// liste + reports (campaign) + encours (campaign) - accompagnement en cours (wekan)
	var siretsToInsert []core.Siret

	campaignReports := utils.Convert(n.fromCampaignReports, campaignAction.getSiret)
	campaignEnCours := utils.Convert(n.fromCampaignEncours, campaignAction.getSiret)
	siretsToInsert = append(n.fromListe, campaignReports...)
	siretsToInsert = append(siretsToInsert, campaignEnCours...)
	return slices.DeleteFunc(siretsToInsert, deleteSiretsFromSiretsSlice(n.fromWekanAccompagnementEnCoursSirets()))
}

func (n newCampaignSirets) campaignEnCoursToInsert() []campaignAction {
	// encours (campaign) - accompagnement en cours (wekan)
	return slices.DeleteFunc(n.fromCampaignEncours, deleteCampaignActionFromSiretsSlice(n.fromWekanAccompagnementEnCoursSirets()))
}

func (n newCampaignSirets) campaignReportsToInsert() []campaignAction {
	// reports (campaign) - accompagnement en cours (wekan)
	reports := slices.DeleteFunc(n.fromCampaignReports, deleteCampaignActionFromSiretsSlice(n.fromWekanAccompagnementEnCoursSirets()))
	reports = slices.DeleteFunc(reports, deleteCampaignReportDelai3mois)
	reports = utils.Convert(reports, updateCampaignReport)
	return reports
}

func (n newCampaignSirets) wekanAnalyseEnCoursToMutatePasDAccompagnement() []core.Siret {
	// analyse en cours (wekan) - siretToInsert
	return slices.DeleteFunc(n.fromWekanAnalyseEnCoursSirets(), deleteSiretsFromSiretsSlice(n.siretsToInsert()))
}

func newCampaignHandler(kanbanService core.KanbanService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var params newCampaignParams

		err := c.Bind(&params)
		fmt.Println(params)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		err = checkParams(c, params)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		wekanDomainRegexp, err := campaign.GetCampaignWekanDomainRegexp(c, params.FromCampaignID)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		sirets, err := selectSirets(c, wekanDomainRegexp, kanbanService, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		// - création de la nouvelle campagne (titre calculé à partir du nom de la liste)
		campaignID, err := campaign.Create(
			c,
			"Campagne de prise de contact "+params.FromListeDetection,
			wekanDomainRegexp,
			params.DateFin,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		// - insertion des sirets (sauf ceux qui sont accompagnement en cours) ->
		_, err = campaign.AddSirets(c, campaignID, sirets.siretsToInsert(), "signaux.faibles")
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		// - insertion des actions `take` (les gens qui avaient des entreprises en cours de contact les conservent dans la campagne d'après) (sauf sirets qui sont accompagnement en cours)
		err = campaignTakeSiretUnsafe(c, campaignID, sirets.campaignEnCoursToInsert())
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		// - insertion des actions `report` (avec un décallage des reports de 3 mois) (sauf sirets qui sont accompagnement en cours)
		err = campaignReportSiretUnsafe(c, campaignID, sirets.campaignReportsToInsert())
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		// - les cartes Analyse en cours qui ne sont plus dans la campagne suivante passent en `Pas d'accompagnement`
		//      - on récupère toutes les cartes en cours d'analyse (a priori commentées dans la campagne précédentes)
		//      - on retranche les cartes dont les sirets sont encore dans la nouvelle campagne
		//      - on bascule toutes celles qui restent dans la liste «Pas d'accompagnement»

		err = mutatePasDAccompagnement(c, sirets, kanbanService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	}
}

func mutatePasDAccompagnement(ctx context.Context, sirets newCampaignSirets, kanbanService core.KanbanService) error {
	user, found := kanbanService.GetUser("signaux.faibles")
	if !found {
		return errors.New("L'utilisateur 'signaux.faibles' n'a pas été trouvé ?!")
	}
	cardsToMutate := slices.DeleteFunc(sirets.fromWekanAnalyseEnCours, func(card libwekan.Card) bool {
		return slices.Contains(sirets.siretsToInsert(), siretWithNameFunc(sirets.config)(card))
	})
	for _, card := range cardsToMutate {
		err := kanbanService.MoveCardListWithTitle(ctx, card, "Pas d'accompagnement", user)
		if err != nil {
			return err
		}
	}
	return nil
}

func selectSirets(ctx context.Context, wekanDomainRegexp string, kanbanService core.KanbanService, params newCampaignParams) (sirets newCampaignSirets, err error) {
	sirets.fromWekanAccompagnementEnCours, err = selectSiretsFromAccompagnement(ctx, wekanDomainRegexp, kanbanService)
	if err != nil {
		return sirets, err
	}
	slog.Info("sirets provenant des accompagnements wekan", slog.Int("siretsFromAccompagnementEnCours", len(sirets.fromWekanAccompagnementEnCours)))

	sirets.fromListe, err = selectSiretsFromListe(ctx, params)
	if err != nil {
		return sirets, err
	}
	slog.Info("sirets provenant de la liste", slog.Int("nb_sirets", len(sirets.fromListe)))

	sirets.fromCampaignReports, err = selectSiretsFromAction(ctx, params, "withdraw")
	if err != nil {
		return sirets, err
	}
	slog.Info("sirets provenant des reports", slog.Int("siretsFromReport", len(sirets.fromCampaignReports)))

	sirets.fromCampaignEncours, err = selectSiretsFromAction(ctx, params, "take")
	if err != nil {
		return sirets, err
	}
	slog.Info("sirets provenant des en cours", slog.Int("siretsFromEncours", len(sirets.fromCampaignEncours)))

	sirets.fromWekanAnalyseEnCours, err = selectSiretsFromAnalyseEnEcours(ctx, wekanDomainRegexp, kanbanService)
	if err != nil {
		return sirets, err
	}
	slog.Info("sirets provenant des analyses en cours wekan", slog.Int("siretsFromAnalyseEnEcours", len(sirets.fromWekanAnalyseEnCours)))

	sirets.config = kanbanService.GetWekanConfig()

	return sirets, nil
}

func selectSiretsFromAccompagnement(ctx context.Context, wekanDomainRegexp string, kanbanService core.KanbanService) ([]libwekan.Card, error) {
	return kanbanService.SelectCardsFromListeAndDomainRegexp(ctx, wekanDomainRegexp, "Accompagnement en cours")
}

func selectSiretsFromAnalyseEnEcours(ctx context.Context, wekanDomainRegexp string, kanbanService core.KanbanService) ([]libwekan.Card, error) {
	return kanbanService.SelectCardsFromListeAndDomainRegexp(ctx, wekanDomainRegexp, "Analyse en cours")
}

//go:embed sql/siretsFromAction.sql
var sqlSiretsFromAction string

func selectSiretsFromAction(ctx context.Context, params newCampaignParams, actionName string) ([]campaignAction, error) {
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

//go:embed sql/siretsFromListe.sql
var sqlSiretsFromListe string

func selectSiretsFromListe(ctx context.Context, params newCampaignParams) ([]core.Siret, error) {
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

func checkParams(ctx context.Context, params newCampaignParams) error {
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

//go:embed sql/unsafeInsertCampaignTakeSiret.sql
var unsafeInsertCampaignTakeSiret string

func campaignTakeSiretUnsafe(ctx context.Context, campaignID campaign.CampaignID, actions []campaignAction) error {
	conn := db.Get()
	for _, action := range actions {
		_, err := conn.Exec(ctx, unsafeInsertCampaignTakeSiret, action.username, action.siret, campaignID)
		if err != nil {
			return err
		}
	}
	return nil
}

//go:embed sql/unsafeInsertCampaignReportSiret.sql
var unsafeInsertCampaignReportSiret string

func campaignReportSiretUnsafe(ctx context.Context, campaignID campaign.CampaignID, actions []campaignAction) error {
	conn := db.Get()
	for _, action := range actions {
		_, err := conn.Exec(ctx, unsafeInsertCampaignReportSiret, action.username, action.actionDetail, action.siret, campaignID)
		if err != nil {
			return err
		}
	}
	return nil
}
