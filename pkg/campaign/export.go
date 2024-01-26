package campaign

import (
	"bytes"
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"datapi/pkg/utils"
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type Exports struct {
	Etablissements    []*CampaignEtablissement `json:"etablissements"`
	NbTotal           int                      `json:"nbTotal"`
	CampaignName      string                   `json:"campaignName"`
	WekanDomainRegexp string                   `json:"-"`
}

func (p *Exports) Tuple() []interface{} {
	var ce CampaignEtablissement
	p.Etablissements = append(p.Etablissements, &ce)
	return []interface{}{
		&p.NbTotal,
		&p.WekanDomainRegexp,
		&p.CampaignName,
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
		&ce.Username,
	}
}

func (c CampaignEtablissement) toFields() []string {
	var fields []string
	fields = append(fields, fmt.Sprintf("%d", c.ID))
	fields = append(fields, fmt.Sprintf("%d", c.CampaignID))
	fields = append(fields, fmt.Sprintf("%s", c.CodeDepartement))

	fields = append(fields, string(c.Siret))
	fields = append(fields, c.RaisonSociale)

	if c.Action != nil {
		fields = append(fields, *c.Action)
	} else {
		fields = append(fields, "pending")
	}

	if c.Detail != nil {
		fields = append(fields, *c.Detail)
	} else {
		fields = append(fields, "")
	}

	if c.Username != nil {
		fields = append(fields, *c.Username)
	} else {
		fields = append(fields, "non attribué")
	}

	if c.List != nil {
		fields = append(fields, *c.List)
	} else {
		fields = append(fields, "Aucun partage d'information")
	}
	return fields
}

func (e Exports) toCSV() []byte {
	csvBuffer := new(bytes.Buffer)
	writer := csv.NewWriter(csvBuffer)
	writer.Comma = ';'
	headers := []string{
		"id_campaign_etablissement",
		"id_campaign",
		"code_departement",
		"siret",
		"raison_sociale",
		"statut",
		"detail_statut",
		"username",
		"statut_accompagnement",
	}
	writer.Write(headers)
	for _, etablissement := range e.Etablissements {
		writer.Write(etablissement.toFields())
	}
	writer.Flush()
	return csvBuffer.Bytes()
}

func exportHandlerFunc(kanbanService core.KanbanService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var s core.Session
		s.Bind(c)

		campaignID, err := strconv.Atoi(c.Param("campaignID"))
		if err != nil {
			c.JSON(400, `/campaign/actions/taken/:campaignID: le parametre campaignID doit être un entier`)
			return
		}
		boards := kanbanService.SelectBoardsForUsername(libwekan.Username(s.Username))
		exports, err := selectExport(c, CampaignID(campaignID), boards, libwekan.Username(s.Username), kanbanService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		csvBytes := exports.toCSV()
		filename := fmt.Sprintf("export-campaign-%s-%s", exports.CampaignName, time.Now().Format("060102"))
		c.Writer.Header().Set("Content-disposition", "attachment;filename="+slug.Make(filename)+".csv")
		c.Data(http.StatusOK, "text/csv", csvBytes)
	}
}

func selectExport(ctx context.Context, campaignID CampaignID, boards []libwekan.ConfigBoard,
	username libwekan.Username, kanbanService core.KanbanService) (exports Exports, err error) {

	zones := zonesFromBoards(boards)
	err = db.Scan(ctx, &exports, sqlSelectExports, campaignID, zones, username)
	if err != nil {
		return Exports{}, err
	}
	if len(exports.Etablissements) == 0 {
		return Exports{Etablissements: []*CampaignEtablissement{}}, nil
	}

	// limiter les boards scannées au périmètre de la campagne
	re, err := regexp.CompilePOSIX(exports.WekanDomainRegexp)
	if err != nil {
		return exports, err
	}

	matchingBoards := utils.Filter(boards, boardMatchesRegexpFunc(re))
	err = appendCardsToExport(ctx, &exports, matchingBoards, kanbanService, username)
	return exports, err
}

func appendCardsToExport(ctx context.Context, exports *Exports,
	boards []libwekan.ConfigBoard, kanbanService core.KanbanService, username libwekan.Username) error {
	sirets := utils.Convert(exports.Etablissements, func(c *CampaignEtablissement) core.Siret { return c.Siret })
	boardIDs := utils.Convert(boards, func(board libwekan.ConfigBoard) libwekan.BoardID { return board.Board.ID })
	cards, err := kanbanService.SelectCardsFromSiretsAndBoardIDs(ctx, sirets, boardIDs, username)
	if err != nil {
		return err
	}
	for _, etablissement := range exports.Etablissements {
		if card, ok := utils.First(cards, func(card core.KanbanCard) bool { return card.Siret == etablissement.Siret }); ok {
			etablissement.CardID = &card.ID
			etablissement.Description = &card.Description
			etablissement.List = &card.ListTitle
		}
	}
	return nil
}
