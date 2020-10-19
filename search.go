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
	sqlSearch := `
	with summaries as (
		select 
			v.siret, v.siren,	v.raison_sociale, v.commune, v.libelle_departement, v.code_departement,
			v.chiffre_affaire, v.arrete_bilan, v.variation_ca, v.resultat_expl, v.effectif, v.libelle_n5,
			v.libelle_n1, v.code_activite, v.last_procol,	v.activite_partielle,	f.id is not null as followed,
			v.siege, v.raison_sociale, v.territoire_industrie
			from v_summary v
			left join etablissement_follow f on f.siret = v.siret and f.active and f.username = $10
			where (v.siret ilike $6 or v.raison_sociale ilike $7)
			and coalesce(v.libelle_liste, $5) = $5
			and (v.siege or $11)
			order by v.raison_sociale, v.siret
	), perms as (
		select  from sumarries
	),
	select v.hausse_urssaf, v.alert, 
	, count(*) over (),	true as visible, true as in_zone,
	
	from summaries s
	inner join perms p on p.siret = s.siret
	where
	and (v.roles_entreprise && $1 or $8)
	and (v.code_departement=any($2) or $9)
	limit $3 offset $4
		;`

	liste, err := findAllListes()
	if err != nil {
		return searchResult{}, errorToJSON(500, err)
	}
	zoneGeo := params.roles.zoneGeo()
	limit := viper.GetInt("searchPageLength")

	rows, err := db.Query(context.Background(), sqlSearch,

		zoneGeo,
		zoneGeo,
		limit,
		params.page*limit,
		liste[0].ID,
		params.search+"%",
		"%"+params.search+"%",
		params.ignoreRoles,
		params.ignoreZone,
		params.username,
		!params.siegeUniquement,
	)

	if err != nil {
		return searchResult{}, errorToJSON(500, err)
	}

	var search searchResult
	for rows.Next() {
		var r Summary
		err := rows.Scan(&r.Siret,
			&r.Siren,
			&r.RaisonSociale,
			&r.Commune,
			&r.LibelleDepartement,
			&r.Departement,
			&r.Score,
			&r.Diff,
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
			&r.Visible,
			&r.InZone,
			&r.Followed,
			&r.Siege,
			&r.Groupe,
			&r.TerrInd,
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
