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

type CampaignEtablissementID int

type CampaignEtablissement struct {
	ID                  CampaignEtablissementID `json:"id"`
	CampaignID          int                     `json:"campaignID"`
	Siret               core.Siret              `json:"siret"`
	Alert               *string                 `json:"alert"`
	Followed            bool                    `json:"followed"`
	FirstAlert          bool                    `json:"firstAlert"`
	EtatAdministratif   string                  `json:"etatAdministratif"`
	CodeDepartement     string                  `json:"codeDepartement"`
	Action              *string                 `json:"action,omitempty"`
	Rank                int                     `json:"rank"`
	RaisonSociale       string                  `json:"raisonSociale"`
	RaisonSocialeGroupe *string                 `json:"raisonSocialeGroupe,omitempty"`
	Username            *string                 `json:"username,omitempty"`
	CardID              *libwekan.CardID        `json:"cardID,omitempty"`
	Description         *string                 `json:"description,omitempty"`
	Detail              *string                 `json:"detail,omitempty"`
}

type Pending struct {
	Etablissements    []*CampaignEtablissement `json:"etablissements"`
	NbTotal           int                      `json:"nbTotal"`
	Page              int                      `json:"page,omitempty"`
	PageMax           int                      `json:"pageMax,omitempty"`
	PageSize          int                      `json:"pageSize,omitempty"`
	WekanDomainRegexp string                   `json:"-"`
}

func pendingHandler(kanbanService core.KanbanService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var s core.Session
		s.Bind(c)
		id, err := strconv.Atoi(c.Param("campaignID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, "`"+c.Param("campaignID")+"` n'est pas un identifiant valide")
			return
		}
		username := libwekan.Username(s.Username)
		boards := kanban.SelectBoardsForUser(username)
		pending, err := selectPending(c, CampaignID(id), boards, core.Page{10, 0}, username, kanbanService)

		if err != nil {
			c.JSON(http.StatusInternalServerError, "erreur inattendue: "+err.Error())
			return
		}
		c.JSON(http.StatusOK, pending)
	}
}

func (p *Pending) Tuple() []interface{} {
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
		&ce.Detail,
	}
}

func selectPending(ctx context.Context, campaignID CampaignID, boards []libwekan.ConfigBoard, page core.Page, username libwekan.Username, kanbanService core.KanbanService) (pending Pending, err error) {
	pending.Etablissements = make([]*CampaignEtablissement, 0)

	zones := zonesFromBoards(boards)
	err = db.Scan(ctx, &pending, sqlSelectPendingEtablissement, campaignID, zones, username)
	if err != nil {
		return Pending{}, err
	}
	if len(pending.Etablissements) == 0 {
		return Pending{Etablissements: []*CampaignEtablissement{}}, nil
	}

	// limiter les boards scannées au périmètre de la campagne
	re, err := regexp.CompilePOSIX(pending.WekanDomainRegexp)
	if err != nil {
		return pending, err
	}
	matchingBoards := utils.Filter(boards, boardMatchesRegexpFunc(re))
	err = appendCardsToPending(ctx, &pending, matchingBoards, kanbanService, username)
	return pending, err
}

func appendCardsToPending(ctx context.Context, pending *Pending, boards []libwekan.ConfigBoard, kanbanService core.KanbanService, username libwekan.Username) error {
	sirets := utils.Convert(pending.Etablissements, func(c *CampaignEtablissement) core.Siret { return c.Siret })
	boardIDs := utils.Convert(boards, func(board libwekan.ConfigBoard) libwekan.BoardID { return board.Board.ID })
	cards, err := kanbanService.SelectCardsFromSiretsAndBoardIDs(ctx, sirets, boardIDs, username)
	if err != nil {
		return err
	}
	for _, etablissement := range pending.Etablissements {
		if card, ok := utils.First(cards, func(card core.KanbanCard) bool { return card.Siret == etablissement.Siret }); ok {
			etablissement.CardID = &card.ID
			etablissement.Description = &card.Description
		}
	}
	return nil
}

func boardMatchesRegexpFunc(re *regexp.Regexp) func(libwekan.ConfigBoard) bool {
	return func(board libwekan.ConfigBoard) bool {
		slug := board.Board.Slug
		return re.MatchString(string(slug))
	}
}
