package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"datapi/pkg/kanban"
	"datapi/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"regexp"
	"strconv"
)

type MyActions struct {
	Etablissements    []*CampaignEtablissement `json:"etablissements"`
	NbTotal           int                      `json:"nbTotal"`
	Page              int                      `json:"page"`
	PageMax           int                      `json:"pageMax"`
	PageSize          int                      `json:"pageSize"`
	WekanDomainRegexp string                   `json:"-"`
}

func (p *MyActions) Tuple() []interface{} {
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
		&ce.CodeDepartement,
	}
}

func myActionsHandler(kanbanService core.KanbanService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var s core.Session
		s.Bind(c)

		campaignID, err := strconv.Atoi(c.Param("campaignID"))
		if err != nil {
			c.JSON(400, `/campaign/myactions/:campaignID: le parametre campaignID doit être un entier`)
			return
		}
		boards := kanban.SelectBoardsForUser(libwekan.Username(s.Username))
		myActions, err := selectMyActions(c, CampaignID(campaignID), boards, libwekan.Username(s.Username), kanbanService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "une erreur est survenue", "detail": err.Error()})
		}
		c.JSON(200, myActions)
	}
}

func selectMyActions(ctx context.Context, campaignID CampaignID, boards []libwekan.ConfigBoard, username libwekan.Username, kanbanService core.KanbanService) (myActions MyActions, err error) {
	zones := zonesFromBoards(boards)
	err = db.Scan(ctx, &myActions, sqlSelectMyActions, campaignID, zones, username)
	if err != nil {
		return MyActions{}, err
	}
	// limiter les boards scannées au périmètre de la campagne
	re, err := regexp.CompilePOSIX(myActions.WekanDomainRegexp)
	if err != nil {
		return myActions, err
	}
	matchingBoards := utils.Filter(boards, boardMatchesRegexpFunc(re))
	err = appendCardsToMyActions(ctx, &myActions, matchingBoards, kanbanService, username)
	return myActions, err
}

func appendCardsToMyActions(ctx context.Context, myActions *MyActions, boards []libwekan.ConfigBoard, kanbanService core.KanbanService, username libwekan.Username) error {
	sirets := utils.Convert(myActions.Etablissements, func(c *CampaignEtablissement) core.Siret { return c.Siret })
	boardIDs := utils.Convert(boards, func(board libwekan.ConfigBoard) libwekan.BoardID { return board.Board.ID })
	cards, err := kanbanService.SelectCardsFromSiretsAndBoardIDs(ctx, sirets, boardIDs, username)
	if err != nil {
		return err
	}
	for _, etablissement := range myActions.Etablissements {
		if card, ok := utils.First(cards, func(card core.KanbanCard) bool { return card.Siret == etablissement.Siret }); ok {
			etablissement.CardID = &card.ID
			etablissement.Description = &card.Description
		}
	}
	return nil
}
