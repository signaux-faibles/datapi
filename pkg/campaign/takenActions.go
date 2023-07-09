package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"datapi/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"regexp"
	"strconv"
)

type TakenActions struct {
	Etablissements    []*CampaignEtablissement `json:"etablissements"`
	NbTotal           int                      `json:"nbTotal"`
	WekanDomainRegexp string                   `json:"-"`
}

func (p *TakenActions) Tuple() []interface{} {
	var ce CampaignEtablissement
	p.Etablissements = append(p.Etablissements, &ce)
	return []interface{}{
		&p.NbTotal,
		&p.WekanDomainRegexp,
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
		&ce.Detail,
	}
}

func takenActionsHandler(kanbanService core.KanbanService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var s core.Session
		s.Bind(c)
		campaignID, err := strconv.Atoi(c.Param("campaignID"))
		if err != nil {
			c.JSON(400, `/campaign/actions/taken/:campaignID: le parametre campaignID doit être un entier`)
			return
		}
		boards := kanbanService.SelectBoardsForUsername(libwekan.Username(s.Username))
		allActions, err := selectTakenActions(c, CampaignID(campaignID), boards, libwekan.Username(s.Username), kanbanService)

		if err != nil {
			c.JSON(500, err.Error())
		} else {
			c.JSON(200, allActions)
		}
	}
}

func selectTakenActions(ctx context.Context, campaignID CampaignID, boards []libwekan.ConfigBoard,
	username libwekan.Username, kanbanService core.KanbanService) (takenActions TakenActions, err error) {

	zones := zonesFromBoards(boards)
	err = db.Scan(ctx, &takenActions, sqlSelectTakenActions, campaignID, zones, username)
	if err != nil {
		return TakenActions{}, err
	}
	if len(takenActions.Etablissements) == 0 {
		return TakenActions{Etablissements: []*CampaignEtablissement{}}, nil
	}
	// limiter les boards scannées au périmètre de la campagne
	re, err := regexp.CompilePOSIX(takenActions.WekanDomainRegexp)
	if err != nil {
		return takenActions, err
	}
	matchingBoards := utils.Filter(boards, boardMatchesRegexpFunc(re))
	err = appendCardsToTakenActions(ctx, &takenActions, matchingBoards, kanbanService, username)
	return takenActions, err
}

func appendCardsToTakenActions(ctx context.Context, takenActions *TakenActions,
	boards []libwekan.ConfigBoard, kanbanService core.KanbanService, username libwekan.Username) error {
	sirets := utils.Convert(takenActions.Etablissements, func(c *CampaignEtablissement) core.Siret { return c.Siret })
	boardIDs := utils.Convert(boards, func(board libwekan.ConfigBoard) libwekan.BoardID { return board.Board.ID })
	cards, err := kanbanService.SelectCardsFromSiretsAndBoardIDs(ctx, sirets, boardIDs, username)
	if err != nil {
		return err
	}
	for _, etablissement := range takenActions.Etablissements {
		if card, ok := utils.First(cards, func(card core.KanbanCard) bool { return card.Siret == etablissement.Siret }); ok {
			etablissement.CardID = &card.ID
			etablissement.Description = &card.Description
		}
	}
	return nil
}
