package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type searchParams struct {
	Search          string   `json:"search"`
	Page            int      `json:"page"`
	Departements    []string `json:"departements,omitempty"`
	Activites       []string `json:"activites,omitempty"`
	EffectifMin     *int     `json:"effectifMin"`
	SiegeUniquement bool     `json:"siegeUniquement"`
	IgnoreRoles     bool     `json:"ignoreRoles"`
	IgnoreZone      bool     `json:"ignoreZone"`
	username        string
	roles           scope
}

type searchResult struct {
	From    int        `json:"from"`
	To      int        `json:"to"`
	Total   int        `json:"total"`
	NBF1    int        `json:"nbF1"`
	NBF2    int        `json:"nbF2"`
	PageMax int        `json:"pageMax"`
	Page    int        `json:"page"`
	Results []*Summary `json:"results"`
}

func searchEtablissementHandler(c *gin.Context) {
	var params searchParams

	err := c.Bind(&params)
	if err != nil {
		c.Abort()
		return
	}

	params.username = c.GetString("username")

	if len(params.Search) < 3 {
		c.JSON(400, "search string length < 3")
		return
	}

	if params.Page < 0 {
		c.JSON(400, "page has to be integer >= 0")
		return
	}

	params.roles = scopeFromContext(c)

	result, Jerr := searchEtablissement(params)
	if Jerr != nil {
		c.JSON(Jerr.Code(), Jerr.Error())
		return
	}
	c.JSON(200, result)

}

func searchEtablissement(params searchParams) (searchResult, Jerror) {
	liste, err := findAllListes()
	if err != nil {
		return searchResult{}, errorToJSON(500, err)
	}
	zoneGeo := params.roles.zoneGeo()
	limit := viper.GetInt("searchPageLength")

	offset := params.Page * limit

	summaryparams := summaryParams{
		zoneGeo, &limit, &offset, &liste[0].ID, &params.Search, &params.IgnoreRoles, &params.IgnoreZone,
		params.username, params.SiegeUniquement, "raison_sociale", &False, nil, params.Departements, nil, params.EffectifMin, nil, nil, params.Activites,
	}

	summaries, err := getSummaries(summaryparams)
	if err != nil {
		return searchResult{}, errorToJSON(500, err)
	}

	var search searchResult

	search.Results = summaries.summaries
	if summaries.global.count != nil {
		search.Total = *summaries.global.count
		search.NBF1 = *summaries.global.countF1
		search.NBF2 = *summaries.global.countF2
	}
	search.From = limit*params.Page + 1
	search.To = limit*params.Page + len(search.Results)
	search.Page = params.Page
	search.PageMax = (search.Total - 1) / limit

	if len(search.Results) == 0 {
		return searchResult{}, newJSONerror(204, "empty page")
	}

	return search, nil
}
