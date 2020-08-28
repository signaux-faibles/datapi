package main

import (
	"context"
	"database/sql"
	"log"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func getAllEntreprises(c *gin.Context) {
	// log.Println("getAllEntreprises")
	// db := c.MustGet("DB").(*sql.DB)
	// entreprises, err := findAllEntreprises(db)
	// if err != nil {
	// 	c.JSON(500, err.Error())
	// }
	// c.JSON(200, entreprises)
}

func getEntreprise(c *gin.Context) {
	roles := scopeFromContext(c)
	siren := c.Param("siren")
	isValidSiret, err := regexp.MatchString("[0-9]{9}", siren)
	if !isValidSiret || err != nil {
		c.JSON(400, "SIREN valide obligatoire")
		return
	}
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

func getAllEtablissements(c *gin.Context) {
	log.Println("getAllEtablissements")
	db := c.MustGet("DB").(*sql.DB)
	etablissements, err := findAllEtablissements(db)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, etablissements)
}

func getEtablissement(c *gin.Context) {
	roles := scopeFromContext(c)
	siret := c.Param("siret")
	isValidSiret, err := regexp.MatchString("[0-9]{14}", siret)
	if !isValidSiret || err != nil {
		c.JSON(400, "SIRET valide obligatoire")
	}
	var etablissements Etablissements
	etablissements.Query.Sirets = []string{siret}
	err = etablissements.load(roles)

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
	isValidSiret, err := regexp.MatchString("[0-9]{9}", siren)
	if !isValidSiret || err != nil {
		c.JSON(400, "SIREN valide obligatoire")
		return
	}
	var etablissements Etablissements
	etablissements.Query.Sirens = []string{siren}
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

func followEntreprise(c *gin.Context) {
	log.Println("followEntreprise")
	db := c.MustGet("DB").(*sql.DB)
	siren := c.Param("siren")
	// TODO: valider SIREN
	if siren == "" {
		c.JSON(400, "SIREN obligatoire")
		return
	}
	userID := ""
	follow := Follow{UserID: userID, Siren: siren}
	err := follow.createEntrepriseFollow(db)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, follow)
}

func getEntrepriseComments(c *gin.Context) {
	log.Println("getEntreprisesComments")
	db := c.MustGet("DB").(*sql.DB)
	siren := c.Param("siren")
	// TODO: valider SIREN
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
	// TODO: valider SIREN
	if siren == "" {
		c.JSON(400, "SIREN obligatoire")
		return
	}
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
	err = prepareImport(&tx)

	sourceEntreprise := viper.GetString("sourceEntreprise")
	log.Printf("processing entreprise file %s", sourceEntreprise)
	err = processEntreprise(sourceEntreprise, &tx)
	if err != nil {
		c.AbortWithError(500, err)
	}

	sourceEtablissement := viper.GetString("sourceEtablissement")
	log.Printf("processing etablissement file %s", sourceEtablissement)
	err = processEtablissement(sourceEtablissement, &tx)
	if err != nil {
		c.AbortWithError(500, err)
	}

	err = upgradeVersion(&tx)
	if err != nil {
		c.AbortWithError(500, err)
	}

	tx.Commit(context.Background())
}
