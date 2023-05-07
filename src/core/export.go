package core

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
)

//func getExportSiret(s session, siret string) (Card, error) {
//	var exports KanbanDBExports
//	exports, exportsFields := exports.NewDbExport()
//	err := db.Get().QueryRow(context.Background(), sqlDbExportSingle, s.roles, s.Username, siret).Scan(exportsFields...)
//	if err != nil {
//		return Card{}, err
//	}
//
//	var c = Card{
//		dbExport: exports[0],
//	}
//	if s.hasRole("wekan") {
//		wekanCards, err := selectWekanCardsFromSiret(s.Username, siret)
//		if err != nil {
//			return Card{}, err
//		}
//		c.WekanCards = wekanCards
//	}
//
//	return c, nil
//}

//// TODO: factoriser avec getCards
//func getExport(ctx context.Context, s session, params KanbanSelectCardsForUserParams) (Cards, error) {
//	var cards Cards
//	var cardsMap = make(map[string]*Card)
//	var sirets []string
//	var followedSirets []string
//	wcu := oldWekanConfig.forUser(s.Username)
//	userID := oldWekanConfig.userID(s.Username)
//
//	if _, ok := oldWekanConfig.Users[s.Username]; s.hasRole("wekan") && params.Type != "no-card" && ok {
//		// Export wekan + DB
//		//var username *string
//		//if params.Type == "my-cards" {
//		//	username = &s.Username
//		//}
//		//boardIds := wcu.boardIds()
//		//swimlaneIds := wcu.swimlaneIdsForZone(params.Zone)
//		//listIds := wcu.listIdsForStatuts(params.Statut)
//		//labelIds := wcu.labelIdsForLabels(params.Labels)
//		//labelMode := params.LabelMode
//
//		wekanCards, err := kanban.SelectFollowsForUser()
//		//wekanCards, err := selectWekanCards(username, boardIds, swimlaneIds, listIds, labelIds, labelMode, params.Since)
//		if err != nil {
//			return nil, err
//		}
//		for _, w := range wekanCards {
//			siret, err := w.Siret()
//			if err != nil {
//				continue
//			}
//			card := Card{nil, []*WekanCard{w}, nil}
//			cards = append(cards, &card)
//			if _, ok := cardsMap[siret]; !ok {
//				cardsMap[siret] = &card
//				sirets = append(sirets, siret)
//			} else {
//				c := cardsMap[siret].WekanCards
//				c = append(c, card.WekanCards...)
//				cardsMap[siret].WekanCards = c
//			}
//
//			if utils.Contains(append(w.Members, w.Assignees...), userID) {
//				followedSirets = append(followedSirets, siret)
//			}
//		}
//		var exports KanbanDBExports
//		if params.Type == "my-cards" {
//			sirets = followedSirets
//		}
//		var cursor pgx.Rows
//		cursor, err = db.Get().Query(context.Background(), sqlDbExport, s.roles, s.Username, sirets)
//
//		if err != nil {
//			return nil, err
//		}
//		for cursor.Next() {
//			var s []interface{}
//			exports, s = exports.NewDbExport()
//			err := cursor.Scan(s...)
//			if err != nil {
//				return nil, err
//			}
//		}
//		cursor.Close()
//		for _, s := range exports {
//			card := cardsMap[s.Siret]
//			if card == nil {
//				card = &Card{}
//			}
//			card.KanbanDBExport = s
//			cardsMap[s.Siret] = card
//		}
//	} else {
//		// export DB uniquement
//		boardIds := wcu.boardIds()
//		var wekanCards []*WekanCard
//		var err error
//
//		if s.hasRole("wekan") {
//			wekanCards, err = selectWekanCards(&s.Username, boardIds, nil, nil, nil, false, nil)
//			if err != nil {
//				return nil, err
//			}
//		}
//
//		var excludeSirets = make(map[string]struct{})
//		for _, w := range wekanCards {
//			if s.hasRole("wekan") {
//				siret, err := w.Siret()
//				if err != nil {
//					continue
//				}
//				excludeSirets[siret] = struct{}{}
//			}
//		}
//		var exports KanbanDBExports
//		var cursor pgx.Rows
//		cursor, err = db.Get().Query(
//			context.Background(),
//			sqlDbExportFollow,
//			s.roles,
//			s.Username,
//			params.Zone,
//		)
//
//		if err != nil {
//			return nil, err
//		}
//		for cursor.Next() {
//			var s []interface{}
//			exports, s = exports.NewDbExport()
//			err := cursor.Scan(s...)
//			if err != nil {
//				return nil, err
//			}
//		}
//		cursor.Close()
//		for _, s := range exports {
//			if _, ok := excludeSirets[s.Siret]; !ok {
//				card := Card{nil, nil, s}
//				cards = append(cards, &card)
//			}
//		}
//	}
//	return cards.dbExportsOnly(), nil
//}

func (cards KanbanExports) xlsx(wekan bool) ([]byte, error) {
	xlFile := xlsx.NewFile()
	xlSheet, err := xlFile.AddSheet("extract")
	if err != nil {
		return nil, utils.ErrorToJSON(http.StatusInternalServerError, err)
	}

	row := xlSheet.AddRow()
	row.AddCell().Value = "Raison sociale"
	row.AddCell().Value = "Siret"
	row.AddCell().Value = "Type d'établissement"
	row.AddCell().Value = "Tête de groupe"
	row.AddCell().Value = "Département"
	row.AddCell().Value = "Commune"
	row.AddCell().Value = "Territoire d'industrie"
	row.AddCell().Value = "Secteur d'activité"
	row.AddCell().Value = "Activité"
	row.AddCell().Value = "Secteurs COVID-19"
	row.AddCell().Value = "Statut juridique"
	row.AddCell().Value = "Date d'ouverture"
	row.AddCell().Value = "Création de l'entreprise"
	row.AddCell().Value = "Effectif Entreprise"
	row.AddCell().Value = "Activité Partielle"
	row.AddCell().Value = "Dette sociale"
	row.AddCell().Value = "Part salariale"
	row.AddCell().Value = "Année Exercice"
	row.AddCell().Value = "Chiffre d'affaires"
	row.AddCell().Value = "Excédent brut d'exploitation"
	row.AddCell().Value = "Résultat d'exploitation"
	row.AddCell().Value = "Procédure collective"
	row.AddCell().Value = "Détection Signaux Faibles"
	row.AddCell().Value = "Début du suivi"
	if wekan {
		row.AddCell().Value = "Étiquettes"
		row.AddCell().Value = "Présentation de l'enteprise, difficultés et actions"
		row.AddCell().Value = "Fin du suivi Wekan"
		row.AddCell().Value = "Tableau"
		row.AddCell().Value = "Dernière modification Wekan"
		row.AddCell().Value = "Archive"
	}

	for _, e := range cards {
		row := xlSheet.AddRow()
		row.AddCell().Value = e.RaisonSociale
		row.AddCell().Value = e.Siret
		row.AddCell().Value = e.TypeEtablissement
		row.AddCell().Value = e.TeteDeGroupe
		row.AddCell().Value = e.Departement
		row.AddCell().Value = e.Commune
		row.AddCell().Value = e.TerritoireIndustrie
		row.AddCell().Value = e.SecteurActivite
		row.AddCell().Value = e.Activite
		row.AddCell().Value = e.SecteursCovid
		row.AddCell().Value = e.StatutJuridique
		row.AddCell().Value = e.DateOuvertureEtablissement
		row.AddCell().Value = e.DateCreationEntreprise
		row.AddCell().Value = e.Effectif
		row.AddCell().Value = e.ActivitePartielle
		row.AddCell().Value = e.DetteSociale
		row.AddCell().Value = e.PartSalariale
		row.AddCell().Value = e.AnneeExercice
		row.AddCell().Value = e.ChiffreAffaire
		row.AddCell().Value = e.ExcedentBrutExploitation
		row.AddCell().Value = e.ResultatExploitation
		row.AddCell().Value = e.ProcedureCollective
		row.AddCell().Value = e.DetectionSF
		row.AddCell().Value = e.DateDebutSuivi
		if wekan {
			row.AddCell().Value = strings.Join(e.Labels, ", ")
			row.AddCell().Value = e.DescriptionWekan
			row.AddCell().Value = e.DateFinSuivi
			row.AddCell().Value = e.Board
			row.AddCell().Value = e.LastActivity.Format("02/01/2006")
			row.AddCell().Value = boolToString(e.Archived, "oui", "non")
		}
	}
	data := bytes.NewBuffer(nil)
	file := bufio.NewWriter(data)
	xlFile.Write(file)
	file.Flush()
	return data.Bytes(), nil
}

func boolToString(boolean bool, true string, false string) string {
	if boolean {
		return true
	}
	return false
}

func (c KanbanExport) docx(head ExportHeader) (Docx, error) {
	data, err := json.MarshalIndent(c, " ", " ")
	script := viper.GetString("docxifyPath")
	dir := viper.GetString("docxifyWorkingDir")
	python := viper.GetString("docxifyPython")
	if err != nil {
		return Docx{}, err
	}
	fmt.Println(string(data))
	cmd := exec.Command(python, script, head.Auteur, head.Date.Format("02/01/2006"), "Haute")
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(data)
	var out bytes.Buffer
	var outErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &outErr
	err = cmd.Run()
	if err != nil {

		fmt.Println(string(outErr.Bytes()))
		return Docx{}, err
	}
	file := out.Bytes()
	if len(file) == 0 {
		fmt.Println("no docx file generated")
	}
	if len(outErr.Bytes()) > 0 {
		fmt.Println(outErr.String())
	}
	return Docx{
		filename: fmt.Sprintf("export-%s-%s.docx", strings.Replace(c.RaisonSociale, " ", "-", -1), c.Siret),
		data:     file,
	}, nil
}

func getEtablissementsFollowedByCurrentUser(c *gin.Context) {
	username := c.GetString("username")
	scope := scopeFromContext(c)
	follow := Follow{Username: &username}
	follows, err := follow.list(scope)
	if err != nil {
		c.JSON(err.Code(), err.Error())
		return
	}
	c.JSON(200, follows)
}

func getXLSXFollowedByCurrentUser(c *gin.Context) {
	var s session
	s.Bind(c)
	var params KanbanSelectCardsForUserParams
	c.Bind(&params)

	var ok bool
	params.User, ok = kanban.GetUser(libwekan.Username(s.Username))
	if !ok {
		c.JSON(http.StatusForbidden, "le nom d'utilisateur n'est pas reconnu")
		return
	}

	exports, err := kanban.ExportFollowsForUser(c, params, db.Get(), s.roles)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	xlsx, err := exports.xlsx(s.hasRole("wekan"))
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	filename := fmt.Sprintf("export-suivi-%s.xlsx", time.Now().Format("060102"))
	c.Writer.Header().Set("Content-disposition", "attachment;filename="+filename)
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", xlsx)
}

type Docx struct {
	filename string
	data     []byte
}

type Docxs []Docx

func (docxs Docxs) zip() []byte {
	var zipData bytes.Buffer
	zipFile := zip.NewWriter(&zipData)
	for _, docx := range docxs {
		w, _ := zipFile.CreateHeader(&zip.FileHeader{
			Name:     docx.filename,
			Modified: time.Now(),
			Method:   0,
		})
		w.Write(docx.data)
	}
	zipFile.Close()
	return zipData.Bytes()
}

func getDOCXFollowedByCurrentUser(c *gin.Context) {
	var s session
	s.Bind(c)
	var params KanbanSelectCardsForUserParams
	c.Bind(&params)

	var ok bool
	params.User, ok = kanban.GetUser(libwekan.Username(s.Username))
	if !ok {
		c.JSON(http.StatusForbidden, "le nom d'utilisateur n'est pas reconnu")
		return
	}

	exports, err := kanban.ExportFollowsForUser(c, params, db.Get(), s.roles)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}

	header := ExportHeader{
		Auteur: s.auteur,
		Date:   time.Now(),
	}

	var docxs Docxs
	for _, export := range exports {
		docx, err := export.docx(header)
		if err != nil {
			utils.AbortWithError(c, err)
			return
		}
		docxs = append(docxs, docx)
	}
	filename := fmt.Sprintf("export-suivi-%s.zip", time.Now().Format("060102"))
	c.Writer.Header().Set("Content-disposition", "attachment;filename="+filename)
	c.Data(200, "application/zip", docxs.zip())
}

func getDOCXFromSiret(c *gin.Context) {
	var s session
	s.Bind(c)
	//
	//siret := c.Param("siret")
	//card, err := getExportSiret(s, siret)
	//if err != nil {
	//	utils.AbortWithError(c, err)
	//	return
	//}
	//
	//header := ExportHeader{
	//	Auteur: s.auteur,
	//	Date:   time.Now(),
	//}
	//docx, err := card.docx(header)
	//if err != nil {
	//	utils.AbortWithError(c, err)
	//	return
	//}
	//
	//c.Writer.Header().Set("Content-disposition", "attachment;filename="+docx.filename)
	//c.Data(200, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docx.data)
}
