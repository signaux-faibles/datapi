package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func getEntreprise(c *gin.Context) {
	roles := scopeFromContext(c)
	siren := c.Param("siren")
	siret, err := getSiegeFromSiren(siren)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var etablissements Etablissements
	etablissements.Query.Sirets = []string{siret}
	err = etablissements.load(roles)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	entreprise, ok := etablissements.Entreprises[siren]
	if !ok {
		c.JSON(404, "ressource non disponible")
		return
	}
	for _, v := range etablissements.Etablissements {
		entreprise.Etablissements = append(entreprise.Etablissements, v)
	}
	c.JSON(200, entreprise)
}

func getEtablissement(c *gin.Context) {
	roles := scopeFromContext(c)
	siret := c.Param("siret")
	var etablissements Etablissements
	etablissements.Query.Sirets = []string{siret}
	err := etablissements.load(roles)

	result, ok := etablissements.Etablissements[siret]
	if !ok {
		c.JSON(404, "etablissement non disponible")
		return
	}
	entreprise := etablissements.Entreprises[siret[0:9]]
	result.Entreprise = &entreprise

	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, result)
}

func getEntrepriseEtablissements(c *gin.Context) {
	roles := scopeFromContext(c)
	siren := c.Param("siren")
	var etablissements Etablissements
	etablissements.Query.Sirens = []string{siren}
	err := etablissements.load(roles)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	entreprise, ok := etablissements.Entreprises[siren]
	if !ok {
		c.JSON(404, "ressource non disponible")
		return
	}

	for _, v := range etablissements.Etablissements {
		entreprise.Etablissements = append(entreprise.Etablissements, v)
	}

	c.JSON(200, entreprise)
}

func getEntreprisesFollowedByUser(c *gin.Context) {
	log.Println("getEntreprisesFollowedByUser")
	db := c.MustGet("DB").(*sql.DB)
	userID := ""
	follow := Follow{UserID: userID}
	entreprises, err := follow.findEntreprisesFollowedByUser(db)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, entreprises)
}

func followEtablissement(c *gin.Context) {
	siren := c.Param("siren")
	fmt.Println(siren)
	// // TODO: valider SIREN
	// if siren == "" {
	// 	c.JSON(400, "SIREN obligatoire")
	// 	return
	// }
	// userID := ""
	// follow := Follow{UserID: userID, Siren: siren}
	// err := follow.createEntrepriseFollow(db)
	// if err != nil {
	// 	c.JSON(500, err.Error())
	// }
	// c.JSON(200, follow)
}

func getEntrepriseComments(c *gin.Context) {
	log.Println("getEntreprisesComments")
	db := c.MustGet("DB").(*sql.DB)
	siren := c.Param("siren")
	if siren == "" {
		c.JSON(400, "SIREN obligatoire")
		return
	}
	entreprise := Entreprise{Siren: siren}
	comments, err := entreprise.findCommentsBySiren(db)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, comments)
}

func addEntrepriseComment(c *gin.Context) {
	log.Println("addEntrepriseComment")
	db := c.MustGet("DB").(*sql.DB)
	siren := c.Param("siren")
	message := c.Param("message")
	if message == "" {
		c.JSON(400, "Message obligatoire")
		return
	}
	userID := ""
	comment := Comment{Siren: siren, UserID: userID, Message: message}
	err := comment.createEntrepriseComment(db)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, comment)
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

	err = liste.getScores(roles)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, liste)
}

func getListeScores(c *gin.Context) {
	roles := scopeFromContext(c)

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

	err = liste.getScores(roles)

	if err != nil {
		if err.Error() == "no rows in result set" {
			c.AbortWithStatus(204)
			return
		}
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, liste)
}

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

func importHandler(c *gin.Context) {
	tx, err := db.Begin(context.Background())
	if err != nil {
		c.AbortWithError(500, err)
	}
	log.Print("preparing database for import")
	err = prepareImport(&tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	sourceEntreprise := viper.GetString("sourceEntreprise")
	log.Printf("processing entreprise file %s", sourceEntreprise)
	err = processEntreprise(sourceEntreprise, &tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	sourceEtablissement := viper.GetString("sourceEtablissement")
	log.Printf("processing etablissement file %s", sourceEtablissement)
	err = processEtablissement(sourceEtablissement, &tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	log.Printf("upgrading version of objects")
	err = upgradeVersion(&tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	log.Print("refreshing materialized views")
	err = refreshMaterializedViews(&tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	log.Print("commiting changes to database")
	tx.Commit(context.Background())

	log.Print("drop dead data")
	_, err = db.Exec(context.Background(), "vacuum full;")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
}

func validSiren(c *gin.Context) {
	siren := c.Param("siren")
	match, err := regexp.MatchString("[0-9]{9}", siren)
	if err != nil || !match {
		c.AbortWithStatusJSON(400, "SIREN valide obligatoire")
	}
	c.Next()
}

func validSiret(c *gin.Context) {
	siren := c.Param("siret")
	match, err := regexp.MatchString("[0-9]{14}", siren)
	if err != nil || !match {
		c.AbortWithStatusJSON(400, "SIRET valide obligatoire")
	}
	c.Next()
}
