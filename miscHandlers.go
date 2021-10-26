package main

import (
	"context"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

func getCodesNaf(c *gin.Context) {
	rows, err := db.Query(context.Background(), "select code, libelle from naf where niveau=1")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	var naf = make(map[string]string)
	for rows.Next() {
		var code string
		var libelle string
		err := rows.Scan(&code, &libelle)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		naf[code] = libelle
	}
	c.JSON(200, naf)
}

func getDepartements(c *gin.Context) {
	rows, err := db.Query(context.Background(), "select code, libelle from departements")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	var departements = make(map[string]string)
	for rows.Next() {
		var code string
		var libelle string
		err := rows.Scan(&code, &libelle)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		departements[code] = libelle
	}
	c.JSON(200, departements)
}

func getRegions(c *gin.Context) {
	rows, err := db.Query(context.Background(), `select r.libelle, array_agg(d.code order by d.code) from regions r
	inner join departements d on d.id_region = r.id
	group by r.libelle`)

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	var regions = make(map[string][]string)
	for rows.Next() {
		var region string
		var departements []string
		err := rows.Scan(&region, &departements)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		regions[region] = departements
	}
	c.JSON(200, regions)
}

func validSiren(c *gin.Context) {
	siren := c.Param("siren")
	match, err := regexp.MatchString("^[0-9]{9}$", siren)
	if err != nil || !match {
		c.AbortWithStatusJSON(400, "SIREN valide obligatoire")
	}
	c.Next()
}

func validSiret(c *gin.Context) {
	siren := c.Param("siret")
	match, err := regexp.MatchString("^[0-9]{14}$", siren)
	if err != nil || !match {
		c.AbortWithStatusJSON(400, "SIRET valide obligatoire")
	}
	c.Next()
}

func coalescepString(pointers ...*string) *string {
	for _, i := range pointers {
		if i != nil {
			return i
		}
	}
	return nil
}

func coalescepTime(pointers ...*time.Time) *time.Time {
	for _, i := range pointers {
		if i != nil {
			return i
		}
	}
	return nil
}

type session struct {
	username string
	auteur   string
	roles    scope
}

func (s *session) bind(c *gin.Context) {
	s.username = c.GetString("username")
	s.auteur = c.GetString("given_name") + " " + c.GetString("family_name")
	s.roles = scopeFromContext(c)
}

func (s session) hasRole(role string) bool {
	return contains(s.roles, role)
}
