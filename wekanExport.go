package main

import (
	"context"
	"time"
)

// WekanExport fournit les champs nécessaires pour l'export Wekan
type WekanExport struct {
	Departement               string
	Commune                   string
	TerritoireIndustrie       string
	SecteurActivite           string
	ActiviteLibelle           string
	ActiviteCode              string
	SecteurCovid              string
	StatutJuridique           string
	DateCreationEtablissement time.Time
	DateCreationEntreprise    time.Time
	DernierEffectif           string
	DernierEffectifDate       string
	ActivitePartielle         bool
	AugmentationDetteSociale  bool
	DetteSalariale            bool
	ExerciceDiane             int
	ChiffreAffaire            int
	ExcedentBrutExploitation  int
	ResultatExploitation      int
	ProcedureCollective       string
	Detection                 string
}

type abstractProcol struct {
	DateEffet time.Time `json:"date_effet"`
	Action    string    `json:"action"`
	Stade     string    `json:"stade"`
}

type dbFollowExport struct {
	codeDepartement                   string
	libelleDepartement                string
	commune                           string
	codeTerritoireIndustrie           string
	libelleTerritoireIndustrie        string
	statutJuridiqueN1                 string
	statutJuridiqueN2                 string
	statutJuridiqueN3                 string
	dateCreationEtablissement         time.Time
	dateCreationEntreprise            time.Time
	dernierEffectif                   int
	dateDernierEffectif               time.Time
	dateArreteBilan                   time.Time
	ExerciceDiane                     int
	ChiffreAffaire                    float64
	ChiffreAffairePrecedent           float64
	VariationCA                       float64
	ResultatExploitation              float64
	ResultatExploitationPrecedent     float64
	ExcedentBrutExploitation          float64
	ExcedentBrutExploitationPrecedent float64
	HistoriqueProcedureCollective     []abstractProcol
	CurrentListe                      string
	PremiereListe                     string
	DerniereListe                     string
	DerniereAlerte                    string
}

func getWekanExport(sirets []string) ([]dbFollowExport, error) {
	var exports []dbFollowExport
	sqlExport := `select 
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
			aet.last_alert
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
		where et.siret = any($1)
	`

	rows, err := db.Query(context.Background(), sqlExport, sirets)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e dbFollowExport
		err := rows.Scan(&e.codeDepartement, &e.libelleDepartement, &e.commune, &e.codeTerritoireIndustrie, &e.libelleTerritoireIndustrie, &e.statutJuridiqueN3, &e.statutJuridiqueN2, &e.statutJuridiqueN2,
			&e.dateCreationEtablissement, &e.dateCreationEntreprise, &e.dernierEffectif, &e.dateDernierEffectif, &e.dateArreteBilan,
			&e.ExerciceDiane, &e.ChiffreAffaire, &e.ChiffreAffairePrecedent, &e.VariationCA, &e.ResultatExploitation, &e.ResultatExploitationPrecedent, &e.ExcedentBrutExploitation, &e.ExcedentBrutExploitationPrecedent,
			&e.HistoriqueProcedureCollective, &e.CurrentListe, &e.DerniereListe, &e.DerniereAlerte)
		if err != nil {
			return nil, err
		}
		exports = append(exports, e)
	} // secteurCovid.get("coucou")
	return exports, nil
}
