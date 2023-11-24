package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"regexp"
)

func upsertCardHandler(kanbanService core.KanbanService) func(*gin.Context) {
	return func(c *gin.Context) {
		var s core.Session
		s.Bind(c)

		type Params struct {
			Description             string                  `json:"description"`
			CampaignEtablissementID CampaignEtablissementID `json:"campaignEtablissementID"`
		}

		var params Params
		err := c.Bind(&params)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "/campaign/card/:campaignEtablissementID doit être un nombre entier"})
			return
		}

		_, message, err := upsertCard(c, params.CampaignEtablissementID, params.Description, kanbanService, libwekan.Username(s.Username))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "une erreur est survenue", "detail": err.Error()})
			return
		}

		stream.Message <- message
		c.Status(http.StatusOK)
	}
}

func upsertCard(ctx context.Context, campaignEtablissementID CampaignEtablissementID,
	description string, kanbanService core.KanbanService, username libwekan.Username) (core.KanbanCard, Message, error) {
	pool := db.Get()
	siret, wekanDomainRegexp, codeDepartement, campaignID, err := getCampaignEtablissement(ctx, campaignEtablissementID, pool)
	if err != nil {
		return core.KanbanCard{}, Message{}, err
	}
	config := kanbanService.LoadConfigForUser(username)
	swimlane, err := selectSwimlane(config, wekanDomainRegexp, codeDepartement)
	if err != nil {
		return core.KanbanCard{}, Message{}, err
	}
	cards, err := kanbanService.SelectCardsFromSiretsAndBoardIDs(ctx, []core.Siret{siret}, []libwekan.BoardID{swimlane.BoardID}, username)
	if len(cards) == 0 {
		params := core.KanbanNewCardParams{
			SwimlaneID:  swimlane.SwimlaneID,
			Description: description,
			Labels:      []libwekan.BoardLabelName{},
			Siret:       siret,
		}
		kanbanCard, err := kanbanService.CreateCard(ctx, params, "signaux.faibles", nil, pool)

		message := Message{
			CampaignEtablissementID: &campaignEtablissementID,
			CampaignID:              campaignID,
			Zone:                    []string{string(codeDepartement)},
			Type:                    "edit-card",
			Username:                string(username),
		}
		return kanbanCard, message, err
	}
	err = kanbanService.UpdateCard(ctx, cards[0], description, username)
	if err != nil {
		return core.KanbanCard{}, Message{}, err
	}
	message := Message{
		CampaignEtablissementID: &campaignEtablissementID,
		CampaignID:              campaignID,
		Zone:                    []string{string(codeDepartement)},
		Type:                    "edit-card",
		Username:                string(username),
	}

	return cards[0], message, err
}

func getCampaignEtablissement(ctx context.Context, campaignEtablissementID CampaignEtablissementID, pool *pgxpool.Pool) (core.Siret, *regexp.Regexp, core.CodeDepartement, CampaignID, error) {
	var siret core.Siret
	var wekanDomainRegexpString string
	var codeDepartement core.CodeDepartement
	var campaignID CampaignID
	err := pool.QueryRow(ctx, sqlSelectCampaignEtablissementID, campaignEtablissementID).Scan(&siret, &wekanDomainRegexpString, &codeDepartement, &campaignID)
	if err != nil {
		return "", nil, "", 0, CampaignEtablissementNotFoundError{err}
	}
	wekanDomainRegexp, err := regexp.CompilePOSIX(wekanDomainRegexpString)
	if err != nil {
		return "", nil, "", 0, CampaignEtablissementNotFoundError{err}
	}
	return siret, wekanDomainRegexp, codeDepartement, campaignID, nil
}

func selectSwimlane(config core.KanbanConfig, wekanDomainRegexp *regexp.Regexp, codeDepartement core.CodeDepartement) (core.KanbanBoardSwimlane, error) {
	swimlanes, _ := config.Departements[codeDepartement]
	for _, swimlane := range swimlanes {
		board := config.Boards[swimlane.BoardID]
		if wekanDomainRegexp.MatchString(string(board.Slug)) {
			return swimlane, nil
		}
	}
	return core.KanbanBoardSwimlane{}, errors.New("aucun couloir disponible pour cette zone")
}
