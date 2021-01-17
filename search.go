package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type searchParams struct {
	search          string
	page            int
	ignoreRoles     bool
	ignoreZone      bool
	siegeUniquement bool
	roles           scope
	username        string
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
	var err error

	params.username = c.GetString("username")

	if search := c.Param("search"); len(search) >= 3 {
		params.search = search
	} else {
		c.JSON(400, "search string length < 3")
		return
	}

	if page, ok := c.GetQuery("page"); ok {
		params.page, err = strconv.Atoi(page)
		if err != nil {
			c.JSON(400, "page has to be integer >= 0")
			return
		}
	}

	if ignoreRoles, ok := c.GetQuery("ignoreroles"); ok {
		params.ignoreRoles, err = strconv.ParseBool(ignoreRoles)
		if err != nil {
			c.JSON(400, "france is either `true` or `false`")
			return
		}
	}

	if ignoreZone, ok := c.GetQuery("ignorezone"); ok {
		params.ignoreZone, err = strconv.ParseBool(ignoreZone)
		if err != nil {
			c.JSON(400, "roles is either  `true` or `false`")
			return
		}
	}

	if siegeUniquement, ok := c.GetQuery("siegeUniquement"); ok {
		params.siegeUniquement, err = strconv.ParseBool(siegeUniquement)
		if err != nil {
			c.JSON(400, "siegeUniquement is either `true` or `false`")
			return
		}
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

	offset := params.page * limit

	summaryparams := summaryParams{
		zoneGeo, &limit, &offset, &liste[0].ID, &params.search, &params.ignoreRoles, &params.ignoreZone,
		params.username, params.siegeUniquement, "raison_sociale", &False, nil, nil, nil, nil, nil, nil, nil,
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
	search.From = limit*params.page + 1
	search.To = limit*params.page + len(search.Results)
	search.Page = params.page
	search.PageMax = (search.Total - 1) / limit

	if len(search.Results) == 0 {
		return searchResult{}, newJSONerror(204, "empty page")
	}

	return search, nil
}
