package core

import (
	"context"
	"datapi/pkg/db"
	"fmt"
	"time"
)

// Summary score + élements sur l'établissement
type Summary struct {
	Siren                       string             `json:"siren"`
	Siret                       string             `json:"siret"`
	ValeurScore                 *float64           `json:"-"`
	DetailScore                 map[string]float64 `json:"-"`
	Diff                        *float64           `json:"-"`
	RaisonSociale               *string            `json:"raison_sociale"`
	Commune                     *string            `json:"commune"`
	LibelleActivite             *string            `json:"libelle_activite"`
	LibelleActiviteN1           *string            `json:"libelle_activite_n1"`
	CodeActivite                *string            `json:"code_activite"`
	CodeDepartement             *string            `json:"departement"`
	LibelleDepartement          *string            `json:"libelleDepartement"`
	Effectif                    *float64           `json:"dernier_effectif"`
	EffectifEntreprise          *float64           `json:"dernier_effectif_entreprise"`
	MontantDetteUrssaf          *float64           `json:"montantDetteUrssaf,omitempty"`
	HausseUrssaf                *bool              `json:"urssaf,omitempty"`
	DetteUrssaf                 *float64           `json:"detteUrssaf,omitempty"`
	ActivitePartielle           *bool              `json:"activite_partielle,omitempty"`
	APHeureConsommeAVG12m       *int               `json:"apHeureConsommeAVG12m,omitempty"`
	APMontantAVG12m             *int               `json:"apMontantAVG12m,omitempty"`
	ChiffreAffaire              *float64           `json:"ca,omitempty"`
	VariationCA                 *float64           `json:"variation_ca,omitempty"`
	ArreteBilan                 *time.Time         `json:"arrete_bilan,omitempty"`
	ExerciceDiane               *int               `json:"exerciceDiane,omitempty"`
	ResultatExploitation        *float64           `json:"resultat_expl,omitempty"`
	ExcedentBrutDExploitation   *float64           `json:"excedent_brut_d_exploitation,omitempty"`
	EtatProcol                  *string            `json:"etat_procol,omitempty"`
	Alert                       *string            `json:"alert,omitempty"`
	Visible                     *bool              `json:"visible,omitempty"`
	InZone                      *bool              `json:"inZone,omitempty"`
	Followed                    *bool              `json:"followed,omitempty"`
	FollowedEntreprise          *bool              `json:"followedEntreprise,omitempty"`
	FirstAlert                  *bool              `json:"firstAlert"`
	Siege                       *bool              `json:"siege"`
	Groupe                      *string            `json:"groupe,omitempty"`
	TerrInd                     *bool              `json:"territoireIndustrie,omitempty"`
	PermUrssaf                  *bool              `json:"permUrssaf,omitempty"`
	PermDGEFP                   *bool              `json:"permDGEFP,omitempty"`
	PermScore                   *bool              `json:"permScore,omitempty"`
	PermBDF                     *bool              `json:"permBDF,omitempty"`
	Comment                     *string            `json:"-"`
	Category                    *string            `json:"-"`
	Since                       *time.Time         `json:"-"`
	SecteurCovid                *string            `json:"secteurCovid,omitempty"`
	EtatAdministratif           *string            `json:"etatAdministratif,omitempty"`
	EtatAdministratifEntreprise *string            `json:"etatAdministratifEntreprise,omitempty"`
	HasDelai                    *bool              `json:"hasDelai,omitempty"`
	Cards                       []KanbanCard       `json:"cards,omitempty"`
}

type Summaries struct {
	//summaryParams summaryParams
	Global struct {
		Count    *int `json:"count"`
		CountF1  *int `json:"countF1"`
		CountF2  *int `json:"countF2"`
		comment  *string
		category *string
		since    *string
	} `json:"stats"`
	Summaries []*Summary `json:"summaries"`
}

func (summaries *Summaries) NewSummary() []interface{} {
	var s Summary
	summaries.Summaries = append(summaries.Summaries, &s)
	t := []interface{}{
		&s.Siret,
		&s.Siren,
		&s.RaisonSociale,
		&s.Commune,
		&s.LibelleDepartement,
		&s.CodeDepartement,
		&s.ValeurScore,
		&s.DetailScore,
		&s.FirstAlert,
		&s.ChiffreAffaire,
		&s.ArreteBilan,
		&s.ExerciceDiane,
		&s.VariationCA,
		&s.ResultatExploitation,
		&s.Effectif,
		&s.EffectifEntreprise,
		&s.LibelleActivite,
		&s.LibelleActiviteN1,
		&s.CodeActivite,
		&s.EtatProcol,
		&s.ActivitePartielle,
		&s.APHeureConsommeAVG12m,
		&s.APMontantAVG12m,
		&s.HausseUrssaf,
		&s.DetteUrssaf,
		&s.Alert,
		&summaries.Global.Count,
		&summaries.Global.CountF1,
		&summaries.Global.CountF2,
		&s.Visible,
		&s.InZone,
		&s.Followed,
		&s.FollowedEntreprise,
		&s.Siege,
		&s.Groupe,
		&s.TerrInd,
		&s.Comment,
		&s.Category,
		&s.Since,
		&s.PermUrssaf,
		&s.PermDGEFP,
		&s.PermScore,
		&s.PermBDF,
		&s.SecteurCovid,
		&s.ExcedentBrutDExploitation,
		&s.EtatAdministratif,
		&s.EtatAdministratifEntreprise,
		&s.HasDelai,
	}
	return t
}

type summaryParams struct {
	zoneGeo               []string
	limit                 *int
	offset                *int
	libelleListe          *string
	currentListe          bool
	filter                *string
	ignoreRoles           *bool
	ignoreZone            *bool
	userName              string
	siegeUniquement       bool
	orderBy               string
	alertOnly             *bool
	etatsProcol           []string
	departements          []string
	suivi                 *bool
	effectifMin           *int
	effectifMax           *int
	sirens                []string
	activites             []string
	effectifMinEntreprise *int
	effectifMaxEntreprise *int
	caMin                 *int
	caMax                 *int
	excludeSecteursCovid  []string
	etatAdministratif     *string
	creationDateThreshold *string
	firstAlert            *bool
	hasntDelai            *bool
	codefiListOnly        *bool
}

func (p summaryParams) toSQLParams() []interface{} {
	var expressionSiret *string
	var expressionRaisonSociale *string

	if p.filter != nil {
		eSiret := *p.filter + "%"
		eRaisonSociale := "%" + *p.filter + "%"
		expressionSiret = &eSiret
		expressionRaisonSociale = &eRaisonSociale
	}
	return []interface{}{
		p.zoneGeo,
		p.limit,
		p.offset,
		p.libelleListe,
		expressionSiret,
		expressionRaisonSociale,
		p.ignoreRoles,
		p.ignoreZone,
		p.userName,
		p.siegeUniquement,
		p.orderBy,
		p.alertOnly,
		p.etatsProcol,
		p.departements,
		p.suivi,
		p.effectifMin,
		p.effectifMax,
		p.sirens,
		p.activites,
		p.effectifMinEntreprise,
		p.effectifMaxEntreprise,
		p.caMin,
		p.caMax,
		p.excludeSecteursCovid,
		p.etatAdministratif,
		p.hasntDelai,
	}
}

// TODO: réécrire cette fonction avec des endpoints séparés (à ne pas oublier pendant le refactor du modèle de donnée)
func getSummaries(params summaryParams) (Summaries, error) {
	var sql string
	var sqlParams []interface{}

	if params.orderBy == "score" {
		if params.currentListe {
			sqlParams = params.toSQLCurrentScoreParams()
			sql = buildSQLCurrentScoreQuery(params.codefiListOnly != nil && *params.codefiListOnly)
		} else {
			sqlParams = params.toSQLScoreParams()
			sql = sqlScore
		}
	} else if params.orderBy == "raison_sociale" {
		p := params.toSQLParams()
		sqlParams = append(p[0:10], p[13], p[15], p[18])
		sqlParams = append(sqlParams, p[19:25]...)
		sql = `select *, null::bool from get_search($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'raison_sociale', null, null, $11, null, $12, null, null, $13, $14, $15, $16, $17, $18, $19) as raison_sociale;`
	} else if params.orderBy == "follow" {
		p := params.toSQLParams()
		sqlParams = append(sqlParams, p[0], p[3], p[8])
		sql = "select *, null::bool from get_follow($1, null, null, $2, null, null, true, true, $3, false, 'follow', false, null, null, true, null, null, null, null, null, null, null, null, null, null) as follow;"
	} else {
		// p := params.toSQLParams()
		// sqlParams = p[0:17]
		// sql = fmt.Sprintf("select * from get_summary($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22) as %s;", params.orderBy)
		return Summaries{}, fmt.Errorf("not implemented: orderBy=%s", params.orderBy)
	}

	rows, err := db.Get().Query(context.Background(), sql, sqlParams...)
	defer rows.Close()
	if err != nil {
		fmt.Println(err)
		return Summaries{}, err
	}

	sms := Summaries{}
	for rows.Next() {
		s := sms.NewSummary()
		err := rows.Scan(s...)
		if err != nil {
			return Summaries{}, err
		}
	}
	return sms, err
}
