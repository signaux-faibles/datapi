package core

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/utils"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
)

// ExportHeader contient l'entête d'un document d'export
type ExportHeader struct {
	Auteur string
	Date   time.Time
}

// WekanExports array of WekanExport
type WekanExports []WekanExport

// WekanExport fournit les champs nécessaires pour l'export Wekan
type WekanExport struct {
	RaisonSociale              string    `json:"raison_sociale"`
	Siret                      string    `json:"siret"`
	TypeEtablissement          string    `json:"type_etablissement"`
	TeteDeGroupe               string    `json:"tete_de_groupe"`
	Departement                string    `json:"departement"`
	Commune                    string    `json:"commune"`
	TerritoireIndustrie        string    `json:"territoire_industrie"`
	SecteurActivite            string    `json:"secteur_activite"`
	Activite                   string    `json:"activite"`
	SecteursCovid              string    `json:"secteurs_covid"`
	StatutJuridique            string    `json:"statut_juridique"`
	DateOuvertureEtablissement string    `json:"date_ouverture_etablissement"`
	DateCreationEntreprise     string    `json:"date_creation_entreprise"`
	Effectif                   string    `json:"effectif"`
	ActivitePartielle          string    `json:"activite_partielle"`
	DetteSociale               string    `json:"dette_sociale"`
	PartSalariale              string    `json:"part_salariale"`
	AnneeExercice              string    `json:"annee_exercice"`
	ChiffreAffaire             string    `json:"ca"`
	ExcedentBrutExploitation   string    `json:"ebe"`
	ResultatExploitation       string    `json:"rex"`
	ProcedureCollective        string    `json:"procol"`
	DetectionSF                string    `json:"detection_sf"`
	DateDebutSuivi             string    `json:"date_debut_suivi"`
	DateFinSuivi               string    `json:"date_fin_suivi"`
	DescriptionWekan           string    `json:"description_wekan"`
	Labels                     []string  `json:"labels"`
	Board                      string    `json:"-"`
	LastActivity               time.Time `json:"lastActivity"`
}

type dbExport struct {
	Siret                             string    `json:"siret"`
	RaisonSociale                     string    `json:"raisonSociale"`
	CodeDepartement                   string    `json:"codeDepartement"`
	LibelleDepartement                string    `json:"libelleDepartement"`
	Commune                           string    `json:"commune"`
	CodeTerritoireIndustrie           string    `json:"codeTerritoireIndustrie"`
	LibelleTerritoireIndustrie        string    `json:"libelleTerritoireIndustrie"`
	Siege                             bool      `json:"siege"`
	TeteDeGroupe                      string    `json:"teteDeGroupe"`
	CodeActivite                      string    `json:"codeActivite"`
	LibelleActivite                   string    `json:"libelleActivite"`
	SecteurActivite                   string    `json:"secteurActivite"`
	StatutJuridiqueN1                 string    `json:"statutJuridiqueN1"`
	StatutJuridiqueN2                 string    `json:"statutJuridiqueN2"`
	StatutJuridiqueN3                 string    `json:"statutJuridiqueN3"`
	DateOuvertureEtablissement        time.Time `json:"dateOuvertureEtablissement"`
	DateCreationEntreprise            time.Time `json:"dateCreationEntreprise"`
	DernierEffectif                   int       `json:"dernierEffecti"`
	DateDernierEffectif               time.Time `json:"dateDernierEffectif"`
	DateArreteBilan                   time.Time `json:"dateArreteBilan"`
	ExerciceDiane                     int       `json:"exerciceDiane"`
	ChiffreAffaire                    float64   `json:"chiffreAffaire"`
	ChiffreAffairePrecedent           float64   `json:"chiffreAffairePrecedent"`
	VariationCA                       float64   `json:"variationCA"`
	ResultatExploitation              float64   `json:"resultatExploitation"`
	ResultatExploitationPrecedent     float64   `json:"resultatExploitationPrecedent"`
	ExcedentBrutExploitation          float64   `json:"excedentBrutExploitation"`
	ExcedentBrutExploitationPrecedent float64   `json:"excedentBrutExploitationPrecedent"`
	DerniereListe                     string    `json:"derniereListe"`
	DerniereAlerte                    string    `json:"derniereAlerte"`
	ActivitePartielle                 bool      `json:"activitePartielle"`
	DetteSociale                      bool      `json:"detteSociale"`
	PartSalariale                     bool      `json:"partSalariale"`
	DateUrssaf                        time.Time `json:"dateUrssaf"`
	ProcedureCollective               string    `json:"procedureCollective"`
	DateProcedureCollective           time.Time `json:"dateProcedureCollective"`
	DateDebutSuivi                    time.Time `json:"dateDebutSuivi"`
	CommentSuivi                      string    `json:"commentSuivi"`
	InZone                            bool      `json:"inZone"`
}

type dbExports []*dbExport

func (exports dbExports) newDbExport() (dbExports, []interface{}) {
	var e dbExport

	exports = append(exports, &e)
	t := []interface{}{
		&e.Siret,
		&e.RaisonSociale,
		&e.CodeDepartement,
		&e.LibelleDepartement,
		&e.Commune,
		&e.CodeTerritoireIndustrie,
		&e.LibelleTerritoireIndustrie,
		&e.Siege,
		&e.TeteDeGroupe,
		&e.CodeActivite,
		&e.LibelleActivite,
		&e.SecteurActivite,
		&e.StatutJuridiqueN1,
		&e.StatutJuridiqueN2,
		&e.StatutJuridiqueN3,
		&e.DateOuvertureEtablissement,
		&e.DateCreationEntreprise,
		&e.DernierEffectif,
		&e.DateDernierEffectif,
		&e.DateArreteBilan,
		&e.ExerciceDiane,
		&e.ChiffreAffaire,
		&e.ChiffreAffairePrecedent,
		&e.VariationCA,
		&e.ResultatExploitation,
		&e.ResultatExploitationPrecedent,
		&e.ExcedentBrutExploitation,
		&e.ExcedentBrutExploitationPrecedent,
		&e.DerniereListe,
		&e.DerniereAlerte,
		&e.ActivitePartielle,
		&e.DetteSociale,
		&e.PartSalariale,
		&e.DateUrssaf,
		&e.ProcedureCollective,
		&e.DateProcedureCollective,
		&e.DateDebutSuivi,
		&e.CommentSuivi,
		&e.InZone,
	}
	return exports, t
}

func getExportSiret(s session, siret string) (Card, error) {
	var exports dbExports
	exports, exportsFields := exports.newDbExport()
	err := db.Get().QueryRow(context.Background(), sqlDbExportSingle, s.roles, s.Username, siret).Scan(exportsFields...)
	if err != nil {
		return Card{}, err
	}

	var c = Card{
		dbExport: exports[0],
	}
	if s.hasRole("wekan") {
		wekanCards, err := selectWekanCardsFromSiret(s.Username, siret)
		if err != nil {
			return Card{}, err
		}
		c.WekanCards = wekanCards
	}

	return c, nil
}

//// TODO: factoriser avec getCards
//func getExport(s session, params paramsGetCards) (Cards, error) {
//	var cards Cards
//	var cardsMap = make(map[string]*Card)
//	var sirets []string
//	var followedSirets []string
//	wcu := oldWekanConfig.forUser(s.Username)
//	userID := oldWekanConfig.userID(s.Username)
//
//	if _, ok := oldWekanConfig.Users[s.Username]; s.hasRole("wekan") && params.Type != "no-card" && ok {
//		// Export wekan + DB
//		var username *string
//		if params.Type == "my-cards" {
//			username = &s.Username
//		}
//		boardIds := wcu.boardIds()
//		swimlaneIds := wcu.swimlaneIdsForZone(params.Zone)
//		listIds := wcu.listIdsForStatuts(params.Statut)
//		labelIds := wcu.labelIdsForLabels(params.Labels)
//		labelMode := params.LabelMode
//
//		wekanCards, err := selectWekanCards(username, boardIds, swimlaneIds, listIds, labelIds, labelMode, params.Since)
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
//		var exports dbExports
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
//			exports, s = exports.newDbExport()
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
//			card.dbExport = s
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
//		var exports dbExports
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
//			exports, s = exports.newDbExport()
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

func (cards Cards) xlsx(wekan bool) ([]byte, error) {
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
	}

	for _, c := range cards {
		if c.dbExport != nil {
			es := c.join()
			for _, e := range es {
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
				}
			}
		}
	}
	data := bytes.NewBuffer(nil)
	file := bufio.NewWriter(data)
	xlFile.Write(file)
	file.Flush()
	return data.Bytes(), nil
}

func (c Card) docx(head ExportHeader) (Docx, error) {
	var we WekanExports

	we = append(we, c.join()...)

	data, err := json.MarshalIndent(we, " ", " ")
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
		filename: fmt.Sprintf("export-%s-%s.docx", strings.Replace(c.dbExport.RaisonSociale, " ", "-", -1), c.dbExport.Siret),
		data:     file,
	}, nil
}

func (c Card) join() []WekanExport {
	wc := oldWekanConfig.copy()
	apartSwitch := map[bool]string{
		true:  "Demande sur les 12 derniers mois",
		false: "Pas de demande récente",
	}
	urssafSwitch := map[bool]string{
		true:  "Hausse sur les 3 derniers mois (%s)",
		false: "Pas de hausse sur les 3 derniers mois (%s)",
	}
	salarialSwitch := map[bool]string{
		true:  "Dette salariale existante (%s)",
		false: "Aucune dette salariale existante (%s)",
	}
	siegeSwitch := map[bool]string{
		true:  "Siège social",
		false: "Établissement secondaire",
	}
	procolSwitch := map[string]string{
		"in_bonis":          "In bonis",
		"liquidation":       "Liquidation judiciaire",
		"redressement":      "Redressement judiciaire",
		"plan_continuation": "Plan de continuation",
		"sauvegarde":        "Sauvegarde",
		"plan_sauvegarde":   "Plan de Sauvegarde",
	}

	output := []WekanExport{}
	if len(c.WekanCards) == 0 {
		c.WekanCards = append(c.WekanCards, nil)
	}
	for _, card := range c.WekanCards {
		we := WekanExport{
			RaisonSociale:              c.dbExport.RaisonSociale,
			Siret:                      c.dbExport.Siret,
			TypeEtablissement:          siegeSwitch[c.dbExport.Siege],
			TeteDeGroupe:               c.dbExport.TeteDeGroupe,
			Departement:                fmt.Sprintf("%s (%s)", c.dbExport.LibelleDepartement, c.dbExport.CodeDepartement),
			Commune:                    c.dbExport.Commune,
			TerritoireIndustrie:        c.dbExport.LibelleTerritoireIndustrie,
			SecteurActivite:            c.dbExport.SecteurActivite,
			Activite:                   fmt.Sprintf("%s (%s)", c.dbExport.LibelleActivite, c.dbExport.CodeActivite),
			SecteursCovid:              secteurCovid.get(c.dbExport.CodeActivite),
			StatutJuridique:            c.dbExport.StatutJuridiqueN2,
			DateOuvertureEtablissement: dateCreation(c.dbExport.DateOuvertureEtablissement),
			DateCreationEntreprise:     dateCreation(c.dbExport.DateCreationEntreprise),
			Effectif:                   fmt.Sprintf("%d (%s)", c.dbExport.DernierEffectif, c.dbExport.DateDernierEffectif.Format("01/2006")),
			ActivitePartielle:          apartSwitch[c.dbExport.ActivitePartielle],
			DetteSociale:               fmt.Sprintf(urssafSwitch[c.dbExport.DetteSociale], dateUrssaf(c.dbExport.DateUrssaf)),
			PartSalariale:              fmt.Sprintf(salarialSwitch[c.dbExport.PartSalariale], dateUrssaf(c.dbExport.DateUrssaf)),
			AnneeExercice:              anneeExercice(c.dbExport.ExerciceDiane),
			ChiffreAffaire:             libelleCA(c.dbExport.ChiffreAffaire, c.dbExport.ChiffreAffairePrecedent, c.dbExport.VariationCA),
			ExcedentBrutExploitation:   libelleFin(c.dbExport.ExcedentBrutExploitation),
			ResultatExploitation:       libelleFin(c.dbExport.ResultatExploitation),
			ProcedureCollective:        procolSwitch[c.dbExport.ProcedureCollective],
			DetectionSF:                libelleAlerte(c.dbExport.DerniereListe, c.dbExport.DerniereAlerte),
		}

		if card != nil {
			we.DateDebutSuivi = dateUrssaf(card.StartAt)
			we.DescriptionWekan = strings.TrimSuffix(card.Description+"\n\n"+strings.ReplaceAll(strings.Join(card.Comments, "\n\n"), "#export", ""), "\n")
			we.Labels = wc.labelForLabelsIDs(card.LabelIds, card.BoardId)
			if card.EndAt != nil {
				we.DateFinSuivi = dateUrssaf(*card.EndAt)
			}
			we.Board = card.Board()
			we.LastActivity = card.LastActivity
		} else {
			we.DateDebutSuivi = dateUrssaf(c.dbExport.DateDebutSuivi)
		}
		output = append(output, we)
	}
	return output
}

func anneeExercice(exercice int) string {
	if exercice == 0 {
		return ""
	}
	return fmt.Sprintf("%d", exercice)
}
func dateUrssaf(dt time.Time) string {
	if dt.IsZero() || dt.Format("02/01/2006") == "01/01/1900" || dt.Format("02/01/2006") == "01/01/0001" {
		return "n/c"
	}
	return dt.Format("01/2006")
}
func dateCreation(dt time.Time) string {
	if dt.IsZero() || dt.Format("02/01/2006") == "01/01/1900" || dt.Format("02/01/2006") == "01/01/0001" {
		return "n/c"
	}
	return dt.Format("02/01/2006")
}

func libelleAlerte(liste string, alerte string) string {
	if alerte == "Alerte seuil F1" {
		return fmt.Sprintf("Risque élevé (%s)", liste)
	}
	if alerte == "Alerte seuil F2" {
		return fmt.Sprintf("Risque modéré (%s)", liste)
	}
	if alerte == "Pas d'alerte" {
		return fmt.Sprintf("Pas de risque (%s)", liste)
	}
	return "Hors périmètre"
}

func libelleFin(val float64) string {
	if val == 0 {
		return "n/c"
	}
	return fmt.Sprintf("%.0f k€", val)
}

func libelleCA(val float64, valPrec float64, variationCA float64) string {
	if val == 0 {
		return "n/c"
	}
	if valPrec == 0 {
		return fmt.Sprintf("%.0f k€", val)
	}
	if variationCA > 1.05 {
		return fmt.Sprintf("%.0f k€ (en hausse)", val)
	} else if variationCA < 0.95 {
		return fmt.Sprintf("%.0f k€ (en baisse)", val)
	}
	return fmt.Sprintf("%.0f k€", val)
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

//func getXLSXFollowedByCurrentUser(c *gin.Context) {
//	var s session
//	s.Bind(c)
//	var params paramsGetCards
//	c.Bind(&params)
//
//	export, err := getExport(s, params)
//	if err != nil {
//		utils.AbortWithError(c, err)
//		return
//	}
//
//	xlsx, err := export.xlsx(s.hasRole("wekan"))
//	if err != nil {
//		utils.AbortWithError(c, err)
//		return
//	}
//	filename := fmt.Sprintf("export-suivi-%s.xlsx", time.Now().Format("060102"))
//	c.Writer.Header().Set("Content-disposition", "attachment;filename="+filename)
//	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", xlsx)
//}

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

//func getDOCXFollowedByCurrentUser(c *gin.Context) {
//	var s session
//	s.Bind(c)
//	var params paramsGetCards
//	c.Bind(&params)
//
//	exports, err := getExport(s, params)
//	if err != nil {
//		utils.AbortWithError(c, err)
//		return
//	}
//	header := ExportHeader{
//		Auteur: s.auteur,
//		Date:   time.Now(),
//	}
//
//	var docxs Docxs
//	for _, export := range exports {
//		docx, err := export.docx(header)
//		if err != nil {
//			utils.AbortWithError(c, err)
//			return
//		}
//		docxs = append(docxs, docx)
//	}
//	filename := fmt.Sprintf("export-suivi-%s.zip", time.Now().Format("060102"))
//	c.Writer.Header().Set("Content-disposition", "attachment;filename="+filename)
//	c.Data(200, "application/zip", docxs.zip())
//}

func getDOCXFromSiret(c *gin.Context) {
	var s session
	s.Bind(c)

	siret := c.Param("siret")
	card, err := getExportSiret(s, siret)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}

	header := ExportHeader{
		Auteur: s.auteur,
		Date:   time.Now(),
	}
	docx, err := card.docx(header)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}

	c.Writer.Header().Set("Content-disposition", "attachment;filename="+docx.filename)
	c.Data(200, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docx.data)
}

var sqlDbExport = `select v.siret, v.raison_sociale, v.code_departement, v.libelle_departement, v.commune,
coalesce(v.code_territoire_industrie, ''), coalesce(v.libelle_territoire_industrie, ''), v.siege, coalesce(v.raison_sociale_groupe, ''),
coalesce(v.code_activite, ''), coalesce(v.libelle_n5, 'norme NAF non prise en charge'), coalesce(v.libelle_n1, 'norme NAF non prise en charge'),
v.statut_juridique_n1, v.statut_juridique_n2, v.statut_juridique_n3, coalesce(v.date_ouverture_etablissement, '1900-01-01'),
coalesce(v.date_creation_entreprise, '1900-01-01'), coalesce(v.effectif_entreprise, 0), coalesce(v.date_effectif, '1900-01-01'),
coalesce(v.arrete_bilan, '0001-01-01'), coalesce(v.exercice_diane,0), coalesce(v.chiffre_affaire,0),
coalesce(v.prev_chiffre_affaire,0), coalesce(v.variation_ca, 1), coalesce(v.resultat_expl,0),
coalesce(v.prev_resultat_expl,0), coalesce(v.excedent_brut_d_exploitation,0),
coalesce(v.prev_excedent_brut_d_exploitation,0), coalesce(v.last_list,''), coalesce(v.last_alert,''),
coalesce(v.activite_partielle, false), coalesce(v.hausse_urssaf, false), coalesce(v.presence_part_salariale, false),
coalesce(v.periode_urssaf, '0001-01-01'), v.last_procol, coalesce(v.date_last_procol, '0001-01-01'), coalesce(f.since, '0001-01-01'),
coalesce(f.comment, ''), (permissions($1, v.roles, v.first_list_entreprise, v.code_departement, f.siret is not null)).in_zone
from v_summaries v
left join etablissement_follow f on f.siret = v.siret and f.username = $2 and active
where v.siret = any($3)
order by f.id, v.siret`

var sqlDbExportFollow = `select v.siret, v.raison_sociale, v.code_departement, v.libelle_departement, v.commune,
coalesce(v.code_territoire_industrie, ''), coalesce(v.libelle_territoire_industrie, ''), v.siege, coalesce(v.raison_sociale_groupe, ''),
coalesce(v.code_activite, ''), coalesce(v.libelle_n5, 'norme NAF non prise en charge'), coalesce(v.libelle_n1, 'norme NAF non prise en charge'),
v.statut_juridique_n1, v.statut_juridique_n2, v.statut_juridique_n3, coalesce(v.date_ouverture_etablissement, '1900-01-01'),
coalesce(v.date_creation_entreprise, '1900-01-01'), coalesce(v.effectif_entreprise, 0), coalesce(v.date_effectif, '1900-01-01'),
coalesce(v.arrete_bilan, '0001-01-01'), coalesce(v.exercice_diane,0), coalesce(v.chiffre_affaire,0),
coalesce(v.prev_chiffre_affaire,0), coalesce(v.variation_ca, 1), coalesce(v.resultat_expl,0),
coalesce(v.prev_resultat_expl,0), coalesce(v.excedent_brut_d_exploitation,0),
coalesce(v.prev_excedent_brut_d_exploitation,0), coalesce(v.last_list,''), coalesce(v.last_alert,''),
coalesce(v.activite_partielle, false), coalesce(v.hausse_urssaf, false), coalesce(v.presence_part_salariale, false),
coalesce(v.periode_urssaf, '0001-01-01'), v.last_procol, coalesce(v.date_last_procol, '0001-01-01'), coalesce(f.since, '0001-01-01'),
coalesce(f.comment, ''), (permissions($1, v.roles, v.first_list_entreprise, v.code_departement, f.siret is not null)).in_zone
from v_summaries v
inner join etablissement_follow f on f.siret = v.siret and f.username = $2 and active
where v.code_departement = any($3) or $3 is null
order by f.id, v.siret`

var sqlDbExportSingle = `select v.siret, v.raison_sociale, v.code_departement, v.libelle_departement, v.commune,
coalesce(v.code_territoire_industrie, ''), coalesce(v.libelle_territoire_industrie, ''), v.siege, coalesce(v.raison_sociale_groupe, ''),
coalesce(v.code_activite, ''), coalesce(v.libelle_n5, 'norme NAF non prise en charge'), coalesce(v.libelle_n1, 'norme NAF non prise en charge'),
v.statut_juridique_n1, v.statut_juridique_n2, v.statut_juridique_n3, coalesce(v.date_ouverture_etablissement, '1900-01-01'),
coalesce(v.date_creation_entreprise, '1900-01-01'), coalesce(v.effectif_entreprise, 0), coalesce(v.date_effectif, '1900-01-01'),
coalesce(v.arrete_bilan, '0001-01-01'), coalesce(v.exercice_diane,0), coalesce(v.chiffre_affaire,0),
coalesce(v.prev_chiffre_affaire,0), coalesce(v.variation_ca, 1), coalesce(v.resultat_expl,0),
coalesce(v.prev_resultat_expl,0), coalesce(v.excedent_brut_d_exploitation,0),
coalesce(v.prev_excedent_brut_d_exploitation,0), coalesce(v.last_list,''), coalesce(v.last_alert,''),
coalesce(v.activite_partielle, false), coalesce(v.hausse_urssaf, false), coalesce(v.presence_part_salariale, false),
coalesce(v.periode_urssaf, '0001-01-01'), v.last_procol, coalesce(v.date_last_procol, '0001-01-01'), coalesce(f.since, '0001-01-01'),
coalesce(f.comment, ''), (permissions($1, v.roles, v.first_list_entreprise, v.code_departement, f.siret is not null)).in_zone
from v_summaries v
left join etablissement_follow f on f.siret = v.siret and f.username = $2 and active
where v.siret = $3
order by f.id, v.siret`
