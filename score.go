package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
	ExclureSuivi    *bool    `json:"exclureSuivi"`
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
	Scores  []*Summary        `json:"scores,omitempty"`
	NbF1    int               `json:"nbF1"`
	NbF2    int               `json:"nbF2"`
	From    int               `json:"from"`
	Total   int               `json:"total"`
	Page    int               `json:"page"`
	PageMax int               `json:"pageMax"`
	To      int               `json:"to"`
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
	var suivi *bool
	if liste.Query.ExclureSuivi != nil {
		if *liste.Query.ExclureSuivi {
			s := false
			suivi = &s
		}
	}

	params := summaryParams{
		roles.zoneGeo(), limit, &offset, &liste.ID, &liste.Query.Filter, nil,
		liste.Query.IgnoreZone, username, liste.Query.SiegeUniquement, "score", &True, liste.Query.EtatsProcol,
		liste.Query.Departements, suivi, liste.Query.EffectifMin, liste.Query.EffectifMax, nil,
	}
	summaries, err := getSummaries(params)
	if err != nil {
		return errorToJSON(500, err)
	}

	scores := summaries.summaries
	// TODO: harmoniser les types de sorties pour éviter les remaniements
	if summaries.global.count != nil {
		liste.Total = *summaries.global.count
		liste.NbF1 = *summaries.global.countF1
		liste.NbF2 = *summaries.global.countF2
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
	defer rows.Close()

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
		row.AddCell().Value = fmt.Sprintf("%s", *score.CodeDepartement)
		row.AddCell().Value = fmt.Sprintf("%s", *score.RaisonSociale)
		row.AddCell().Value = fmt.Sprintf("%f", *score.Effectif)
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
