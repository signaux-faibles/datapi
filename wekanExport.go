package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
	"go.mongodb.org/mongo-driver/bson"
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
	RaisonSociale              string `json:"raison_sociale"`
	Siret                      string `json:"siret"`
	TypeEtablissement          string `json:"type_etablissement"`
	TeteDeGroupe               string `json:"tete_de_groupe"`
	Departement                string `json:"departement"`
	Commune                    string `json:"commune"`
	TerritoireIndustrie        string `json:"territoire_industrie"`
	SecteurActivite            string `json:"secteur_activite"`
	Activite                   string `json:"activite"`
	SecteursCovid              string `json:"secteurs_covid"`
	StatutJuridique            string `json:"statut_juridique"`
	DateOuvertureEtablissement string `json:"date_ouverture_etablissement"`
	DateCreationEntreprise     string `json:"date_creation_entreprise"`
	Effectif                   string `json:"effectif"`
	ActivitePartielle          string `json:"activite_partielle"`
	DetteSociale               string `json:"dette_sociale"`
	PartSalariale              string `json:"part_salariale"`
	AnneeExercice              string `json:"annee_exercice"`
	ChiffreAffaire             string `json:"ca"`
	ExcedentBrutExploitation   string `json:"ebe"`
	ResultatExploitation       string `json:"rex"`
	ProcedureCollective        string `json:"procol"`
	DetectionSF                string `json:"detection_sf"`
	DateDebutSuivi             string `json:"date_debut_suivi"`
	DescriptionWekan           string `json:"description_wekan"`
}

type dbExports []dbExport
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

func getExport(roles scope, username string, wekan bool, sirets []string) (WekanExports, error) {
	var wc WekanConfig
	err := wc.load()
	if err != nil {
		return nil, err
	}
	exports, err := getDbExport(roles, username, sirets)
	if err != nil {
		return nil, err
	}
	for _, c := range exports {
		if c.InZone {
			sirets = append(sirets, c.Siret)
		}
	}
	var cards WekanCards
	if wekan {
		cards, err = getDbWekanCards(sirets, wc.siretFields())
		if err != nil {
			return nil, err
		}
	}

	return joinExports(wc, exports, cards), nil
}

func (we WekanExports) xlsx(wekan bool) ([]byte, error) {
	xlFile := xlsx.NewFile()
	xlSheet, err := xlFile.AddSheet("extract")
	if err != nil {
		return nil, errorToJSON(500, err)
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
	row.AddCell().Value = "Effectif"
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
		row.AddCell().Value = "Présentation de l'enteprise, difficultés et actions"
	}

	for _, e := range we {
		row := xlSheet.AddRow()
		if err != nil {
			return nil, errorToJSON(500, err)
		}
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
			row.AddCell().Value = e.DescriptionWekan
		}
	}
	data := bytes.NewBuffer(nil)
	file := bufio.NewWriter(data)
	xlFile.Write(file)
	file.Flush()
	return data.Bytes(), nil
}

func (we WekanExports) docx(head ExportHeader) ([]byte, error) {
	data, err := json.Marshal(we)
	script := viper.GetString("docxifyPath")
	dir := viper.GetString("docxifyWorkingDir")
	python := viper.GetString("docxifyPython")
	if err != nil {
		return nil, err
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
		return nil, err
	}
	file := out.Bytes()
	if len(file) == 0 {
		fmt.Println("no docx file generated")
	}
	if len(outErr.Bytes()) > 0 {
		fmt.Println(outErr.String())
	}
	return file, nil
}

func getDbWekanCards(sirets []string, siretFields []string) (WekanCards, error) {
	if len(sirets) == 0 {
		return nil, nil
	}
	pipeline := []bson.M{
		{
			"$match": bson.M{"type": "cardType-card"},
		}, {
			"$project": bson.M{
				"description":  1,
				"customFields": 1,
				"startAt":      1,
			},
		}, {
			"$unwind": "$customFields",
		}, {
			"$match": bson.M{
				"customFields._id":   bson.M{"$in": siretFields},
				"customFields.value": bson.M{"$in": sirets},
			},
		}, {
			"$project": bson.M{
				"description": 1,
				"startAt":     1,
				"siret":       "$customFields.value",
			},
		},
	}

	cur, err := mgoDB.Collection("cards").Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	var res WekanCards
	err = cur.All(context.Background(), &res)
	return res, err
}

func getDbExport(roles scope, username string, sirets []string) ([]dbExport, error) {
	var exports []dbExport
	sqlExport := `select v.siret, v.raison_sociale, v.code_departement, v.libelle_departement, v.commune,
		coalesce(v.code_territoire_industrie, ''), coalesce(v.libelle_territoire_industrie, ''), v.siege, coalesce(v.raison_sociale_groupe, ''),
		v.code_activite, coalesce(v.libelle_n5, 'norme NAF non prise en charge'), coalesce(v.libelle_n1, 'norme NAF non prise en charge'), 
		v.statut_juridique_n1, v.statut_juridique_n2, v.statut_juridique_n3, coalesce(v.date_ouverture_etablissement, '1900-01-01'), 
		coalesce(v.date_creation_entreprise, '1900-01-01'), coalesce(v.effectif, 0), coalesce(v.date_effectif, '1900-01-01'),
		coalesce(v.arrete_bilan, '0001-01-01'), coalesce(v.exercice_diane,0), coalesce(v.chiffre_affaire,0), 
		coalesce(v.prev_chiffre_affaire,0), coalesce(v.variation_ca, 1), coalesce(v.resultat_expl,0), 
		coalesce(v.prev_resultat_expl,0), coalesce(v.excedent_brut_d_exploitation,0),
		coalesce(v.prev_excedent_brut_d_exploitation,0), coalesce(v.last_list,''), coalesce(v.last_alert,''), 
		coalesce(v.activite_partielle, false), coalesce(v.hausse_urssaf, false), coalesce(v.presence_part_salariale, false), 
		coalesce(v.periode_urssaf, '0001-01-01'), v.last_procol, coalesce(v.date_last_procol, '0001-01-01'), coalesce(f.since, '0001-01-01'),
		(permissions($1, v.roles, v.first_list_entreprise, v.code_departement, f.siret is not null)).in_zone
	from v_summaries v
	left join etablissement_follow f on f.siret = v.siret and f.username = $2 and active
	where v.siret = any($3) or ($3 is null and f.id is not null)
	`

	rows, err := db.Query(context.Background(), sqlExport, roles.zoneGeo(), username, sirets)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e dbExport
		err := rows.Scan(&e.Siret, &e.RaisonSociale, &e.CodeDepartement, &e.LibelleDepartement, &e.Commune,
			&e.CodeTerritoireIndustrie, &e.LibelleTerritoireIndustrie, &e.Siege, &e.TeteDeGroupe,
			&e.CodeActivite, &e.LibelleActivite, &e.SecteurActivite, &e.StatutJuridiqueN1, &e.StatutJuridiqueN2,
			&e.StatutJuridiqueN3, &e.DateOuvertureEtablissement, &e.DateCreationEntreprise, &e.DernierEffectif,
			&e.DateDernierEffectif, &e.DateArreteBilan, &e.ExerciceDiane, &e.ChiffreAffaire, &e.ChiffreAffairePrecedent,
			&e.VariationCA, &e.ResultatExploitation, &e.ResultatExploitationPrecedent, &e.ExcedentBrutExploitation,
			&e.ExcedentBrutExploitationPrecedent, &e.DerniereListe, &e.DerniereAlerte, &e.ActivitePartielle,
			&e.DetteSociale, &e.PartSalariale, &e.DateUrssaf, &e.ProcedureCollective, &e.DateProcedureCollective,
			&e.DateDebutSuivi, &e.InZone)
		if err != nil {
			return nil, err
		}
		exports = append(exports, e)
	}
	return exports, nil
}

func joinExports(wc WekanConfig, exports dbExports, cards WekanCards) WekanExports {
	idx := indexCards(cards)
	var wekanExports []WekanExport

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
	for _, e := range exports {
		we := WekanExport{
			RaisonSociale:              e.RaisonSociale,
			Siret:                      e.Siret,
			TypeEtablissement:          siegeSwitch[e.Siege],
			TeteDeGroupe:               e.TeteDeGroupe,
			Departement:                fmt.Sprintf("%s (%s)", e.LibelleDepartement, e.CodeDepartement),
			Commune:                    e.Commune,
			TerritoireIndustrie:        e.LibelleTerritoireIndustrie,
			SecteurActivite:            e.SecteurActivite,
			Activite:                   fmt.Sprintf("%s (%s)", e.LibelleActivite, e.CodeActivite),
			SecteursCovid:              secteurCovid.get(e.CodeActivite),
			StatutJuridique:            e.StatutJuridiqueN2,
			DateOuvertureEtablissement: dateCreation(e.DateOuvertureEtablissement),
			DateCreationEntreprise:     dateCreation(e.DateCreationEntreprise),
			Effectif:                   fmt.Sprintf("%d (%s)", e.DernierEffectif, e.DateDernierEffectif.Format("01/2006")),
			ActivitePartielle:          apartSwitch[e.ActivitePartielle],
			DetteSociale:               fmt.Sprintf(urssafSwitch[e.DetteSociale], dateUrssaf(e.DateUrssaf)),
			PartSalariale:              fmt.Sprintf(salarialSwitch[e.PartSalariale], dateUrssaf(e.DateUrssaf)),
			AnneeExercice:              anneeExercice(e.ExerciceDiane),
			ChiffreAffaire:             libelleCA(e.ChiffreAffaire, e.ChiffreAffairePrecedent, e.VariationCA),
			ExcedentBrutExploitation:   libelleFin(e.ExcedentBrutExploitation),
			ResultatExploitation:       libelleFin(e.ResultatExploitation),
			ProcedureCollective:        procolSwitch[e.ProcedureCollective],
			DetectionSF:                libelleAlerte(e.DerniereListe, e.DerniereAlerte),
		}

		if i, ok := idx[e.Siret]; ok {
			we.DateDebutSuivi = dateUrssaf(cards[i].StartAt)
			we.DescriptionWekan = cards[i].Description
		} else {
			we.DateDebutSuivi = dateUrssaf(e.DateDebutSuivi)
		}

		wekanExports = append(wekanExports, we)
	}
	return wekanExports
}

func anneeExercice(exercice int) string {
	if exercice == 0 {
		return "n/c"
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

func indexCards(cards WekanCards) map[string]int {
	index := make(map[string]int)

	for i, c := range cards {
		if c.Siret != "" {
			index[c.Siret] = i
		} else {
			log.Printf("no siret for card %s", c.ID)
		}
	}
	return index
}
