package core

import (
	"archive/zip"
	"bufio"
	"bytes"
	"datapi/pkg/db"
	"datapi/pkg/utils"
	"encoding/json"
	"fmt"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
)

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
	err = xlFile.Write(file)
	if err != nil {
		return nil, err
	}
	err = file.Flush()
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func boolToString(boolean bool, true string, false string) string {
	if boolean {
		return true
	}
	return false
}

func (kanbanExport KanbanExport) docx(head ExportHeader) (Docx, error) {
	// docxify prend un tableau en entrée pour un rendu multipage
	input := []interface{}{kanbanExport}
	data, err := json.MarshalIndent(input, " ", " ")
	script := viper.GetString("docxifyPath")
	dir := viper.GetString("docxifyWorkingDir")
	python := viper.GetString("docxifyPython")
	if err != nil {
		return Docx{}, err
	}
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
		filename: fmt.Sprintf("%s-%s-%s.docx",
			strings.Replace(kanbanExport.Board, " ", "-", -1),
			strings.Replace(kanbanExport.RaisonSociale, " ", "-", -1),
			kanbanExport.Siret,
		),
		data: file,
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
	var cards Summaries
	summaries := utils.Convert(follows, followToSummary)
	count := len(summaries)
	cards.Global.Count = &count
	cards.Summaries = summaries
	c.JSON(200, cards)
}

func followToSummary(f Follow) *Summary {
	return f.EtablissementSummary
}

func getXLSXFollowedByCurrentUser(c *gin.Context) {
	var s Session
	s.Bind(c)
	var params KanbanSelectCardsForUserParams
	err := c.Bind(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}
	var ok bool
	params.User, ok = kanban.GetUser(libwekan.Username(s.Username))
	if !ok {
		c.JSON(http.StatusForbidden, "le nom d'utilisateur n'est pas reconnu")
		return
	}
	params.BoardIDs = kanban.ClearBoardIDs(params.BoardIDs, params.User)

	exports, err := kanban.ExportFollowsForUser(c, params, db.Get(), s.Roles)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	xlsxFile, err := exports.xlsx(s.hasRole("wekan"))
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	filename := fmt.Sprintf("export-suivi-%s.xlsx", time.Now().Format("060102"))
	c.Writer.Header().Set("Content-disposition", "attachment;filename="+filename)
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", xlsxFile)
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
		w, err := zipFile.CreateHeader(&zip.FileHeader{
			Name:     docx.filename,
			Modified: time.Now(),
			Method:   0,
		})
		if err != nil {
			return nil
		}
		_, err = w.Write(docx.data)
		if err != nil {
			return nil
		}
	}
	err := zipFile.Close()
	if err != nil {
		return nil
	}
	return zipData.Bytes()
}

func getDOCXFollowedByCurrentUser(c *gin.Context) {
	var s Session
	s.Bind(c)
	var params KanbanSelectCardsForUserParams
	err := c.Bind(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}
	var ok bool
	params.User, ok = kanban.GetUser(libwekan.Username(s.Username))
	if !ok {
		c.JSON(http.StatusForbidden, "le nom d'utilisateur n'est pas reconnu")
		return
	}
	params.BoardIDs = kanban.ClearBoardIDs(params.BoardIDs, params.User)

	exports, err := kanban.ExportFollowsForUser(c, params, db.Get(), s.Roles)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}

	header := ExportHeader{
		Auteur: s.Auteur,
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
	var s Session
	s.Bind(c)

	siret := c.Param("siret")
	kanbanExports, err := kanban.SelectKanbanExportsWithSiret(c, siret, s.Username, db.Get(), s.Roles)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}

	header := ExportHeader{
		Auteur: s.Auteur,
		Date:   time.Now(),
	}

	if len(kanbanExports) > 1 {
		var docxs Docxs
		for _, export := range kanbanExports {
			docx, err := export.docx(header)
			if err != nil {
				utils.AbortWithError(c, err)
				return
			}
			docxs = append(docxs, docx)
		}
		filename := fmt.Sprintf("export-suivi-%s.zip", time.Now().Format("060102"))
		c.Writer.Header().Set("Content-Disposition", "attachment;filename="+filename)
		c.Data(200, "application/zip", docxs.zip())
		return
	}

	docx, err := kanbanExports[0].docx(header)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}

	c.Writer.Header().Set("Content-disposition", "attachment;filename="+docx.filename)
	c.Data(200, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docx.data)
}
