package main

import (
	"context"
	"fmt"
	"time"
)

// Summary score + élements sur l'établissement
type Summary struct {
	Siren                 string             `json:"siren"`
	Siret                 string             `json:"siret"`
	ValeurScore           *float64           `json:"-"`
	DetailScore           map[string]float64 `json:"-"`
	Diff                  *float64           `json:"-"`
	RaisonSociale         *string            `json:"raison_sociale"`
	Commune               *string            `json:"commune"`
	LibelleActivite       *string            `json:"libelle_activite"`
	LibelleActiviteN1     *string            `json:"libelle_activite_n1"`
	CodeActivite          *string            `json:"code_activite"`
	CodeDepartement       *string            `json:"departement"`
	LibelleDepartement    *string            `json:"libelleDepartement"`
	Effectif              *float64           `json:"dernier_effectif"`
	MontantDetteUrssaf    *float64           `json:"montantDetteUrssaf,omitempty"`
	HausseUrssaf          *bool              `json:"urssaf,omitempty"`
	DetteUrssaf           *float64           `json:"detteUrssaf,omitempty"`
	ActivitePartielle     *bool              `json:"activite_partielle,omitempty"`
	APHeureConsommeAVG12m *int               `json:"apHeureConsommeAVG12m,omitempty"`
	APMontantAVG12m       *int               `json:"apMontantAVG12m,omitempty"`
	ChiffreAffaire        *float64           `json:"ca,omitempty"`
	VariationCA           *float64           `json:"variation_ca,omitempty"`
	ArreteBilan           *time.Time         `json:"arrete_bilan,omitempty"`
	ExerciceDiane         *int               `json:"exerciceDiane,omitempty"`
	ResultatExploitation  *float64           `json:"resultat_expl,omitempty"`
	EtatProcol            *string            `json:"etat_procol,omitempty"`
	Alert                 *string            `json:"alert,omitempty"`
	Visible               *bool              `json:"visible,omitempty"`
	InZone                *bool              `json:"inZone,omitempty"`
	Followed              *bool              `json:"followed,omitempty"`
	FollowedEntreprise    *bool              `json:"followedEntreprise,omitempty"`
	FirstAlert            *bool              `json:"firstAlert"`
	Siege                 *bool              `json:"siege"`
	Groupe                *string            `json:"groupe,omitempty"`
	TerrInd               *bool              `json:"territoireIndustrie,omitempty"`
	PermUrssaf            *bool              `json:"permUrssaf,omitempty"`
	PermDGEFP             *bool              `json:"permDGEFP,omitempty"`
	PermScore             *bool              `json:"permScore,omitempty"`
	PermBDF               *bool              `json:"permBDF,omitempty"`
	Comment               *string            `json:"-"`
	Category              *string            `json:"-"`
	Since                 *time.Time         `json:"-"`
}

type summaries struct {
	summaryParams summaryParams
	global        struct {
		count    *int
		countF1  *int
		countF2  *int
		comment  *string
		category *string
		since    *string
	}
	summaries []*Summary
}

func (summaries *summaries) newSummary() []interface{} {
	var s Summary
	summaries.summaries = append(summaries.summaries, &s)
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
		&summaries.global.count,
		&summaries.global.countF1,
		&summaries.global.countF2,
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
	}
	return t
}

type summaryParams struct {
	zoneGeo         []string
	limit           *int
	offset          *int
	libelleListe    *string
	filter          *string
	ignoreRoles     *bool
	ignoreZone      *bool
	userName        string
	siegeUniquement bool
	orderBy         string
	alertOnly       *bool
	etatsProcol     []string
	departements    []string
	suivi           *bool
	effectifMin     *int
	effectifMax     *int
	sirens          []string
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
	}
}

func getSummaries(params summaryParams) (summaries, error) {
	var sql string
	var sqlParams []interface{}

	if params.orderBy == "score" {
		p := params.toSQLParams()
		sqlParams = append(sqlParams, p[0:6]...)
		sqlParams = append(sqlParams, p[7:10]...)
		sqlParams = append(sqlParams, p[12:]...)
		sql = `select * from get_score($1, $2, $3, $4, $5, $6, null, $7, $8, $9, 'score', true, $10, $11, $12, $13, $14, $15) as scores;`
	} else if params.orderBy == "raison_sociale" {
		p := params.toSQLParams()
		sqlParams = p[0:10]
		sql = `select * from get_search($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'raison_sociale', false, null, null, null, null, null, null) as raison_sociale;`
	} else if params.orderBy == "follow" {
		p := params.toSQLParams()
		sqlParams = append(sqlParams, p[0], p[3], p[8])
		sql = "select * from get_follow($1, null, null, $2, null, null, true, true, $3, false, 'follow', false, null, null, true, null, null, null) as follow;"
	} else {
		sqlParams = params.toSQLParams()
		fmt.Println("pouet", params.orderBy)
		sql = fmt.Sprintf("select * from get_summary($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) as %s;", params.orderBy)
	}

	rows, err := db.Query(context.Background(), sql, sqlParams...)
	if err != nil {
		return summaries{}, err
	}

	defer rows.Close()
	sms := summaries{}
	for rows.Next() {
		s := sms.newSummary()
		err := rows.Scan(s...)
		if err != nil {
			return summaries{}, err
		}
	}
	return sms, err
}
