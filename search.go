package main

import (
	"context"
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
	From    int       `json:"from"`
	To      int       `json:"to"`
	Total   int       `json:"total"`
	NBF1    int       `json:"nbF1"`
	NBF2    int       `json:"nbF2"`
	PageMax int       `json:"pageMax"`
	Page    int       `json:"page"`
	Results []Summary `json:"results"`
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
	sqlSearch := `select * from get_summary($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19);`

	liste, err := findAllListes()
	if err != nil {
		return searchResult{}, errorToJSON(500, err)
	}
	zoneGeo := params.roles.zoneGeo()
	limit := viper.GetInt("searchPageLength")

	rows, err := db.Query(context.Background(), sqlSearch,
		zoneGeo,
		limit,
		params.page*limit,
		liste[0].ID,
		params.search+"%",
		"%"+params.search+"%",
		params.ignoreRoles,
		params.ignoreZone,
		params.username,
		params.siegeUniquement,
		"raison_sociale",
		false,
		nil,
		nil,
		nil,
		nil,
		nil,
		false,
		nil,
	)

	if err != nil {
		return searchResult{}, errorToJSON(500, err)
	}

	var search searchResult
	var throwAway interface{}
	for rows.Next() {
		var r Summary
		err := rows.Scan(&r.Siret,
			&r.Siren,
			&r.RaisonSociale,
			&r.Commune,
			&r.LibelleDepartement,
			&r.Departement,
			&r.Score,
			&r.Detail,
			&r.FirstAlert,
			&r.DernierCA,
			&r.ArreteBilan,
			&r.VariationCA,
			&r.DernierREXP,
			&r.DernierEffectif,
			&r.LibelleActivite,
			&r.LibelleActiviteN1,
			&r.CodeActivite,
			&r.EtatProcol,
			&r.ActivitePartielle,
			&r.HausseUrssaf,
			&r.Alert,
			&search.Total,
			&search.NBF1,
			&search.NBF2,
			&r.Visible,
			&r.InZone,
			&r.Followed,
			&r.FollowedEntreprise,
			&r.Siege,
			&r.Groupe,
			&r.TerrInd,
			&throwAway,
			&throwAway,
			&throwAway,
			&r.PermUrssaf,
			&r.PermDGEFP,
			&r.PermScore,
			&r.PermBDF,
		)
		if err != nil {
			return searchResult{}, errorToJSON(500, err)
		}
		search.Results = append(search.Results, r)
	}

	if viper.GetInt("searchPageLength")*params.page >= search.Total {
		return searchResult{}, newJSONerror(204, "empty page")
	}

	search.From = limit*params.page + 1
	search.To = limit*params.page + len(search.Results)
	search.Page = params.page
	search.PageMax = (search.Total - 1) / limit

	return search, nil
}
