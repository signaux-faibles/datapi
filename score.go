package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
)

type paramsListeScores struct {
	Departements    []string `json:"zone,omitempty"`
	EtatsProcol     []string `json:"procol,omitempty"`
	Activites       []string `json:"activite,omitempty"`
	EffectifMin     *int     `json:"effectifMin"`
	EffectifMax     *int     `json:"effectifMax"`
	IgnoreZone      *bool    `json:"ignorezone"`
	ExclureSuivi    bool     `json:"exclureSuivi"`
	SiegeUniquement bool     `json:"siegeUniquement"`
	Page            int      `json:"page"`
	Filter          string   `json:"filter"`
}

// Liste de détection
type Liste struct {
	ID      string            `json:"id"`
	Batch   string            `json:"batch"`
	Algo    string            `json:"algo"`
	Query   paramsListeScores `json:"-"`
	Scores  []Summary         `json:"scores,omitempty"`
	NbF1    int               `json:"nbF1"`
	NbF2    int               `json:"nbF2"`
	From    int               `json:"from"`
	Total   int               `json:"total"`
	Page    int               `json:"page"`
	PageMax int               `json:"pageMax"`
	To      int               `json:"to"`
}

// Summary score + élements sur l'établissement
type Summary struct {
	Siren              string     `json:"siren"`
	Siret              string     `json:"siret"`
	Score              *float64   `json:"-"`
	Diff               *float64   `json:"-"`
	RaisonSociale      *string    `json:"raison_sociale"`
	Commune            *string    `json:"commune"`
	LibelleActivite    *string    `json:"libelle_activite"`
	LibelleActiviteN1  *string    `json:"libelle_activite_n1"`
	CodeActivite       *string    `json:"code_activite"`
	Departement        *string    `json:"departement"`
	LibelleDepartement *string    `json:"libelleDepartement"`
	DernierEffectif    *float64   `json:"dernier_effectif"`
	HausseUrssaf       *bool      `json:"urssaf,omitempty"`
	ActivitePartielle  *bool      `json:"activite_partielle,omitempty"`
	DernierCA          *float64   `json:"ca"`
	VariationCA        *float64   `json:"variation_ca"`
	ArreteBilan        *time.Time `json:"arrete_bilan"`
	DernierREXP        *float64   `json:"resultat_expl"`
	EtatProcol         *string    `json:"etat_procol,omitempty"`
	Alert              *string    `json:"alert,omitempty"`
	Visible            *bool      `json:"visible,omitempty"`
	InZone             *bool      `json:"inZone,omitempty"`
	Followed           *bool      `json:"followed,omitempty"`
	FollowedEntreprise *bool      `json:"followedEntreprise,omitempty"`
	FirstAlert         *bool      `json:"firstAlert"`
	Siege              *bool      `json:"siege"`
	Groupe             *string    `json:"groupe,omitempty"`
	TerrInd            *bool      `json:"territoireIndustrie,omitempty"`
	PermUrssaf         *bool      `json:"permUrssaf,omitempty"`
	PermDGEFP          *bool      `json:"permDGEFP,omitempty"`
	PermScore          *bool      `json:"permScore,omitempty"`
	PermBDF            *bool      `json:"permBDF,omitempty"`
}

func getListes(c *gin.Context) {
	listes, err := findAllListes()
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, listes)
}

func getLastListeScores(c *gin.Context) {
	roles := scopeFromContext(c)
	username := c.GetString("username")
	listes, err := findAllListes()
	if err != nil || len(listes) == 0 {
		c.AbortWithStatus(204)
		return
	}

	var params paramsListeScores
	err = c.Bind(&params)

	if err != nil || len(listes) == 0 {
		c.AbortWithStatus(400)
		return
	}

	liste := Liste{
		ID:    listes[0].ID,
		Query: params,
	}
	limit := viper.GetInt("searchPageLength")
	if limit == 0 {
		c.JSON(418, "searchPageLength must be > 0 in configuration therefore, I'm a teapot.")
		return
	}
	Jerr := liste.getScores(roles, params.Page, &limit, username)
	if Jerr != nil {
		c.JSON(Jerr.Code(), Jerr.Error())
		return
	}
	c.JSON(200, liste)
}

func getXLSListeScores(c *gin.Context) {
	roles := scopeFromContext(c)
	username := c.GetString("username")

	var params paramsListeScores
	err := c.Bind(&params)

	liste := Liste{
		ID:    c.Param("id"),
		Query: params,
	}

	if err != nil || liste.ID == "" {
		c.AbortWithStatus(400)
		return
	}
	limit := viper.GetInt("searchPageLength")
	if limit == 0 {
		c.JSON(418, "searchPageLength must be > 0 in configuration therefore, I'm a teapot.")
		return
	}
	Jerr := liste.getScores(roles, 0, nil, username)

	if Jerr != nil {
		c.JSON(Jerr.Code(), Jerr.Error())
		return
	}

	file, err := liste.toXLS(params)
	c.Writer.Header().Set("Content-disposition", "attachment;filename=extract.xls")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", file)
}

func getListeScores(c *gin.Context) {
	roles := scopeFromContext(c)
	username := c.GetString("username")

	var params paramsListeScores
	err := c.Bind(&params)

	liste := Liste{
		ID:    c.Param("id"),
		Query: params,
	}

	if err != nil || liste.ID == "" {
		c.AbortWithStatus(400)
		return
	}
	limit := viper.GetInt("searchPageLength")
	if limit == 0 {
		c.JSON(418, "searchPageLength must be > 0 in configuration therefore, I'm a teapot.")
		return
	}
	Jerr := liste.getScores(roles, params.Page, &limit, username)

	if Jerr != nil {
		c.JSON(Jerr.Code(), Jerr.Error())
		return
	}

	c.JSON(200, liste)
}

func (liste *Liste) getScores(roles scope, page int, limit *int, username string) Jerror {
	if liste.Batch == "" {
		err := liste.load()
		if err != nil {
			return err
		}
	}
	var offset int
	if limit == nil {
		offset = 0
	} else {
		offset = page * *limit
	}
	sqlScores := `select 
		et.siret,
		et.siren,
		en.raison_sociale, 
		et.commune, 
		d.libelle, 
		d.code,
		s.score,
		s.diff,
		di.chiffre_affaire,
		di.arrete_bilan,
		di.variation_ca,
		di.resultat_expl,
		ef.effectif,
		n.libelle_n5,
		n.libelle_n1,
		et.code_activite,
		coalesce(ep.last_procol, 'in_bonis') as last_procol,
		case when 'dgefp' = any($1) then coalesce(ap.ap, false) else null end as activite_partielle,
		case when 'urssaf' = any($1) then 
			case when u.dette[0] > u.dette[1] or u.dette[1] > u.dette[2] then true else false end 
		else null end as hausseUrssaf,
		s.alert,
		count(case when s.alert='Alerte seuil F1' then 1 else null end) over (),
		count(case when s.alert='Alerte seuil F2' then 1 else null end) over (),
		count(*) over (),
		et.departement=any($1) as inZone,
		f.id is not null as followed,
		r.roles && $1 as visible,
		vs.first_list = $2 as firstAlert,
		et.siege, g.raison_sociale,
		ti.code_commune is not null
	from score0 s
	inner join v_alert_etablissement vs on vs.siret = s.siret
	inner join v_roles r on r.siren = s.siren
	inner join etablissement0 et on et.siret = s.siret
	inner join entreprise0 en on en.siren = s.siren
	inner join departements d on d.code = et.departement
	left join v_naf n on n.code_n5 = et.code_activite
	left join v_last_effectif ef on ef.siret = s.siret
	left join v_hausse_urssaf u on u.siret = s.siret
	left join v_apdemande ap on ap.siret = s.siret
	left join v_last_procol ep on ep.siret = s.siret
	left join v_diane_variation_ca di on di.siren = s.siren
	left join etablissement_follow f on f.siret = s.siret and f.active = true and f.username = $13
	left join entreprise_ellisphere0 g on g.siren = s.siren
	left join terrind ti on ti.code_commune = et.code_commune
	where 
	(r.roles && $1 or f.id is not null)
	and s.libelle_liste = $2
	and (coalesce(ep.last_procol, 'in_bonis')=any($3) or $3 is null)
	and (et.departement=any($4) or $4 is null)
	and (n.code_n1=any($5) or $5 is null)
	and (ef.effectif >= $6 or $6 is null)
	and (ef.effectif <= $7 or $7 is null)
	and (et.departement=any($1) or $8 = true)
	and s.alert != 'Pas d''alerte'
	and (et.siret ilike $11 or en.raison_sociale ilike $12)
	and (f.id is null or $14 = false)
	and (et.siege or $15)
	order by s.score desc, s.siret asc
	limit $9 offset $10;`
	rows, err := db.Query(context.Background(), sqlScores,
		roles.zoneGeo(), liste.ID, liste.Query.EtatsProcol, // $1…
		liste.Query.Departements, liste.Query.Activites, // $4…
		liste.Query.EffectifMin, liste.Query.EffectifMax, // $6…
		liste.Query.IgnoreZone, limit, offset, // $8…
		liste.Query.Filter+"%", "%"+liste.Query.Filter+"%", // $11
		username, liste.Query.ExclureSuivi, !liste.Query.SiegeUniquement, // $13
	)
	if err != nil {
		return errorToJSON(500, err)
	}

	var scores []Summary
	for rows.Next() {
		var score Summary

		err := rows.Scan(
			&score.Siret,
			&score.Siren,
			&score.RaisonSociale,
			&score.Commune,
			&score.LibelleDepartement,
			&score.Departement,
			&score.Score,
			&score.Diff,
			&score.DernierCA,
			&score.ArreteBilan,
			&score.VariationCA,
			&score.DernierREXP,
			&score.DernierEffectif,
			&score.LibelleActivite,
			&score.LibelleActiviteN1,
			&score.CodeActivite,
			&score.EtatProcol,
			&score.ActivitePartielle,
			&score.HausseUrssaf,
			&score.Alert,
			&liste.NbF1,
			&liste.NbF2,
			&liste.Total,
			&score.InZone,
			&score.Followed,
			&score.Visible,
			&score.FirstAlert,
			&score.Siege,
			&score.Groupe,
			&score.TerrInd,
		)
		if err != nil {
			return errorToJSON(500, err)
		}
		scores = append(scores, score)
	}

	if limit == nil {
		i := 1
		limit = &i
	}
	liste.Scores = scores
	liste.Page = page
	liste.PageMax = (liste.Total - 1) / *limit
	liste.From = *limit*page + 1
	liste.To = *limit*page + len(scores)

	if *limit*page > liste.Total || liste.Total == 0 {
		return newJSONerror(204, "empty page")
	}

	return nil
}

func findAllListes() ([]Liste, error) {
	var listes []Liste
	rows, err := db.Query(context.Background(), "select algo, batch, libelle from liste order by batch desc, algo asc")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var l Liste
		err := rows.Scan(&l.Algo, &l.Batch, &l.ID)
		if err != nil {
			return nil, err
		}
		listes = append(listes, l)
	}

	if len(listes) == 0 {
		return nil, errors.New("findAllListes: no list found")
	}
	return listes, nil
}

func (liste *Liste) load() Jerror {
	sqlListe := `select batch, algo from liste where libelle=$1 and version=0`
	row := db.QueryRow(context.Background(), sqlListe, liste.ID)
	batch, algo := "", ""
	err := row.Scan(&batch, &algo)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return newJSONerror(404, "no such list")
		}
		return errorToJSON(500, err)
	}
	liste.Batch = batch
	liste.Algo = algo
	return nil
}

func (liste *Liste) toXLS(params paramsListeScores) ([]byte, Jerror) {
	xlFile := xlsx.NewFile()
	xlSheet, err := xlFile.AddSheet("extract")
	if err != nil {
		return nil, errorToJSON(500, err)
	}

	row := xlSheet.AddRow()
	row.AddCell().Value = "liste"
	row.AddCell().Value = "siren"
	row.AddCell().Value = "siret"
	row.AddCell().Value = "departement"
	row.AddCell().Value = "raison_sociale"
	row.AddCell().Value = "dernier_effectif"
	row.AddCell().Value = "code_activite"
	row.AddCell().Value = "libelle_activite"
	row.AddCell().Value = "alert"

	for _, score := range liste.Scores {
		row := xlSheet.AddRow()
		if err != nil {
			return nil, errorToJSON(500, err)
		}
		row.AddCell().Value = fmt.Sprintf("%s", liste.ID)
		row.AddCell().Value = fmt.Sprintf("%s", score.Siren)
		row.AddCell().Value = fmt.Sprintf("%s", score.Siret)
		row.AddCell().Value = fmt.Sprintf("%s", *score.Departement)
		row.AddCell().Value = fmt.Sprintf("%s", *score.RaisonSociale)
		row.AddCell().Value = fmt.Sprintf("%f", *score.DernierEffectif)
		row.AddCell().Value = fmt.Sprintf("%s", *score.CodeActivite)
		row.AddCell().Value = fmt.Sprintf("%s", *score.LibelleActivite)
		row.AddCell().Value = fmt.Sprintf("%s", *score.Alert)
	}

	sheetParams, err := xlFile.AddSheet("parameters")
	row = sheetParams.AddRow()
	s, err := json.Marshal(params.Activites)
	row.AddCell().Value = "activites"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, err = json.Marshal(params.Departements)
	row.AddCell().Value = "departements"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, err = json.Marshal(params.EffectifMin)
	row.AddCell().Value = "effectifMin"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, err = json.Marshal(params.EffectifMax)
	row.AddCell().Value = "effectifMax"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, err = json.Marshal(params.EtatsProcol)
	row.AddCell().Value = "etatsProcol"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, err = json.Marshal(params.Filter)
	row.AddCell().Value = "filter"
	row.AddCell().Value = string(s)

	row = sheetParams.AddRow()
	s, err = json.Marshal(params.IgnoreZone)
	row.AddCell().Value = "ignorezone"
	row.AddCell().Value = string(s)

	data := bytes.NewBuffer(nil)
	file := bufio.NewWriter(data)
	xlFile.Write(file)
	file.Flush()
	return data.Bytes(), nil
}
