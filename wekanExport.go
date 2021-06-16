package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tealeg/xlsx"
)

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

type abstractProcol struct {
	DateEffet time.Time `json:"date_effet"`
	Action    string    `json:"action"`
	Stade     string    `json:"stade"`
}

type dbExports []dbExport
type dbExport struct {
	Siret                             string           `json:"siret"`
	RaisonSociale                     string           `json:"raisonSociale"`
	CodeDepartement                   string           `json:"codeDepartement"`
	LibelleDepartement                string           `json:"libelleDepartement"`
	Commune                           string           `json:"commune"`
	CodeTerritoireIndustrie           string           `json:"codeTerritoireIndustrie"`
	LibelleTerritoireIndustrie        string           `json:"libelleTerritoireIndustrie"`
	Siege                             bool             `json:"siege"`
	TeteDeGroupe                      string           `json:"teteDeGroupe"`
	CodeActivite                      string           `json:"codeActivite"`
	LibelleActivite                   string           `json:"libelleActivite"`
	SecteurActivite                   string           `json:"secteurActivite"`
	StatutJuridiqueN1                 string           `json:"statutJuridiqueN1"`
	StatutJuridiqueN2                 string           `json:"statutJuridiqueN2"`
	StatutJuridiqueN3                 string           `json:"statutJuridiqueN3"`
	DateOuvertureEtablissement        time.Time        `json:"dateOuvertureEtablissement"`
	DateCreationEntreprise            time.Time        `json:"dateCreationEntreprise"`
	DernierEffectif                   int              `json:"dernierEffecti"`
	DateDernierEffectif               time.Time        `json:"dateDernierEffectif"`
	DateArreteBilan                   time.Time        `json:"dateArreteBilan"`
	ExerciceDiane                     int              `json:"exerciceDiane"`
	ChiffreAffaire                    float64          `json:"chiffreAffaire"`
	ChiffreAffairePrecedent           float64          `json:"chiffreAffairePrecedent"`
	VariationCA                       float64          `json:"variationCA"`
	ResultatExploitation              float64          `json:"resultatExploitation"`
	ResultatExploitationPrecedent     float64          `json:"resultatExploitationPrecedent"`
	ExcedentBrutExploitation          float64          `json:"excedentBrutExploitation"`
	ExcedentBrutExploitationPrecedent float64          `json:"excedentBrutExploitationPrecedent"`
	HistoriqueProcedureCollective     []abstractProcol `json:"historiqueProcedureCollective"`
	DerniereListe                     string           `json:"derniereListe"`
	DerniereAlerte                    string           `json:"derniereAlerte"`
	ActivitePartielle                 bool             `json:"activitePartielle"`
	DetteSociale                      bool             `json:"detteSociale"`
	PartSalariale                     bool             `json:"partSalariale"`
	DateUrssaf                        time.Time        `json:"dateUrssaf"`
	ProcedureCollective               string           `json:"procedureCollective"`
	DateDebutSuivi                    time.Time        `json:"dateDebutSuivi"`
}

func getExport(wekan bool, sirets []string) (WekanExports, error) {
	exports, err := getDbExport(sirets)
	if err != nil {
		return nil, err
	}
	var cards WekanCards
	if wekan {
		cards, err = getDbWekanCards(sirets)
		if err != nil {
			return nil, err
		}
	}
	var wc WekanConfig
	err = wc.load()
	if err != nil {
		return nil, err
	}
	return joinExports(wc, exports, cards), nil
}

func (we WekanExports) xlsx() ([]byte, error) {
	xlFile := xlsx.NewFile()
	xlSheet, err := xlFile.AddSheet("extract")
	if err != nil {
		return nil, errorToJSON(500, err)
	}

	row := xlSheet.AddRow()
	row.AddCell().Value = "Raison sociale"
	row.AddCell().Value = "Siret"
	row.AddCell().Value = "Type Etablissement"
	row.AddCell().Value = "Tête de groupe"
	row.AddCell().Value = "Département"
	row.AddCell().Value = "Commune"
	row.AddCell().Value = "Territoire d'industrie"
	row.AddCell().Value = "Secteur d'activité"
	row.AddCell().Value = "Activité"
	row.AddCell().Value = "Secteur Covid"
	row.AddCell().Value = "Statut Juridique"
	row.AddCell().Value = "Date Ouverture Établissement"
	row.AddCell().Value = "Date Création Entreprise"
	row.AddCell().Value = "Effectif"
	row.AddCell().Value = "Activité Partielle"
	row.AddCell().Value = "Dette Sociale"
	row.AddCell().Value = "Part Salariale"
	row.AddCell().Value = "Année Exercice"
	row.AddCell().Value = "Chiffre Affaire"
	row.AddCell().Value = "E.B.E."
	row.AddCell().Value = "Résultat d'Exploitation"
	row.AddCell().Value = "Procédure Collective"
	row.AddCell().Value = "Détection Signaux-Faibles"
	row.AddCell().Value = "Date Début Suivi"
	row.AddCell().Value = "Description"

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
		row.AddCell().Value = e.DescriptionWekan
	}
	data := bytes.NewBuffer(nil)
	file := bufio.NewWriter(data)
	xlFile.Write(file)
	file.Flush()
	return data.Bytes(), nil
}

func (we WekanExports) docx() ([]byte, error) {
	return nil, nil
}

func getDbWekanCards(sirets []string) (WekanCards, error) {
	return nil, nil
}

func getDbExport(sirets []string) ([]dbExport, error) {
	var exports []dbExport
	sqlExport := `select siret,
			et.departement as code_departement,
			d.libelle as libelle_departement,
			et.commune,
			ti.code_terrind, 
			ti.libelle_terrind,
			cj.libelle, cj2.libelle, cj1.libelle,
			et.creation as creation_etablissement, 
			en.creation as creation_entreprise,
			e.effectif as last_effectif, 
			e.periode as effectif_periode,
			di.arrete_bilan,
			di.exercice_diane,
			di.chiffre_affaire,
			di.prev_chiffre_affaire,
			di.variation_ca,
			di.resultat_expl,
			di.prev_resultat_expl,
			di.excedent_brut_d_exploitation,
			di.prev_excedent_brut_d_exploitation,
			pc.procol_history,
			b.batch as current_list,
			aet.first_list,
			aet.last_list,
			aet.last_alert,
			et.siege,
			et.code_activite,
			ee.raison_sociale as tete_de_groupe,
			n.libelle_n1 as secteur_activite,
			n.libelle_n5 as libelle_activite
		from etablissement0 et
		inner join entreprise0 en on en.siren = et.siren
		inner join v_naf n on n.code_n5 = et.code_activite
		inner join departements d on d.code = et.departement
		left join (select max(batch) as batch from liste) as b on true
		left join v_hausse_urssaf hu on hu.siret = et.siret
		left join categorie_juridique cj on cj.code = en.statut_juridique
		left join categorie_juridique cj2 on substring(cj.code from 0 for 3) = cj2.code
		left join categorie_juridique cj1 on substring(cj.code from 0 for 2) = cj1.code
		left join terrind ti on ti.code_commune = et.code_commune
		left join v_last_effectif e on e.siret = et.siret
		left join v_diane_variation_ca di on di.siren = et.siren
		left join v_last_procol pc on pc.siret = et.siret
		left join v_alert_etablissement aet on aet.siret = et.siret
		left join entreprise_ellisphere0 on ee.siren = et.siren
		where et.siret = any($1)
	`

	rows, err := db.Query(context.Background(), sqlExport, sirets)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e dbExport
		err := rows.Scan(&e.CodeDepartement, &e.LibelleDepartement, &e.Commune, &e.CodeTerritoireIndustrie, &e.LibelleTerritoireIndustrie,
			&e.StatutJuridiqueN3, &e.StatutJuridiqueN2, &e.StatutJuridiqueN2, &e.DateOuvertureEtablissement, &e.DateCreationEntreprise,
			&e.DernierEffectif, &e.DateDernierEffectif, &e.DateArreteBilan, &e.ExerciceDiane, &e.ChiffreAffaire, &e.ChiffreAffairePrecedent,
			&e.VariationCA, &e.ResultatExploitation, &e.ResultatExploitationPrecedent, &e.ExcedentBrutExploitation, &e.ExcedentBrutExploitationPrecedent,
			&e.HistoriqueProcedureCollective, &e.DerniereListe, &e.DerniereAlerte, &e.Siege, &e.CodeActivite, &e.TeteDeGroupe,
			&e.SecteurActivite, &e.LibelleActivite)
		if err != nil {
			return nil, err
		}
		exports = append(exports, e)
	}
	return exports, nil
}

func joinExports(wc WekanConfig, exports dbExports, cards WekanCards) []WekanExport {
	idx := indexCards(cards, wc)
	var wekanExports []WekanExport

	apartSwitch := map[bool]string{
		true:  "Demande sur les 12 derniers mois",
		false: "Pas de demande récente",
	}
	urssafSwitch := map[bool]string{
		true:  "Hausse sur les 3 derniers mois",
		false: "Pas de hausse sur les 3 derniers mois",
	}
	salarialSwitch := map[bool]string{
		true:  "Dette salariale existante %s",
		false: "Aucune dette salariale existante %s",
	}
	siegeSwitch := map[bool]string{
		true:  "Siège",
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
			DateOuvertureEtablissement: e.DateOuvertureEtablissement.Format("02/01/2006"),
			DateCreationEntreprise:     e.DateCreationEntreprise.Format("02/01/2006"),
			Effectif:                   fmt.Sprintf("%d (%s)", e.DernierEffectif, e.DateDernierEffectif.Format("01/2006")),
			ActivitePartielle:          apartSwitch[e.ActivitePartielle],
			DetteSociale:               urssafSwitch[e.DetteSociale],
			PartSalariale:              fmt.Sprintf(salarialSwitch[e.PartSalariale], e.DateUrssaf.Format("01/2006")),
			AnneeExercice:              fmt.Sprintf("%d", e.ExerciceDiane),
			ChiffreAffaire:             libelleFinancier(e.ChiffreAffaire, e.ChiffreAffairePrecedent, 0.05),
			ExcedentBrutExploitation:   fmt.Sprintf("%.0f", e.ExcedentBrutExploitation),
			ResultatExploitation:       fmt.Sprintf("%.0f", e.ResultatExploitation),
			ProcedureCollective:        procolSwitch[e.ProcedureCollective],
			DetectionSF:                libelleAlerte(e.DerniereListe, e.DerniereAlerte),
		}

		if i, ok := idx[e.Siret]; ok {
			we.DateDebutSuivi = cards[i].CreatedAt.Format("02/01/2006")
			we.DescriptionWekan = cards[i].Description
		} else {
			we.DateDebutSuivi = e.DateDebutSuivi.Format("02/01/2006")
		}

		wekanExports = append(wekanExports, we)
	}
	return wekanExports
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

func libelleFinancier(val float64, valPrec float64, threshold float64) string {
	if valPrec == 0 {
		return fmt.Sprintf("%.0f", val)
	}
	if val/valPrec > 1+threshold {
		return fmt.Sprintf("%.0f (en hausse)", val)
	} else if val/valPrec < 1-threshold {
		return fmt.Sprintf("%.0f (en baisse)", val)
	}
	return fmt.Sprintf("%.0f", val)
}

func indexCards(cards WekanCards, wc WekanConfig) map[string]int {
	index := make(map[string]int)
	boards := wc.boards()
	for i, c := range cards {
		var siret string
		siretField := wc.Boards[boards[c.BoardID]].CustomFields.SiretField
		for _, cf := range c.CustomFieds {
			if cf.ID == siretField {
				siret = cf.Value
			}
		}
		if siret != "" {
			index[siret] = i
		} else {
			log.Printf("no siret for card %s", c.ID)
		}
	}
	return index
}
