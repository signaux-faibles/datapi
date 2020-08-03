package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
)

func getAllEntreprises(c *gin.Context) {
	log.Println("getAllEntreprises")
	db := c.MustGet("DB").(*sql.DB)
	entreprises, err := findAllEntreprises(db)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, entreprises)
}

func getEntreprise(c *gin.Context) {
	log.Println("getEntreprise")
	db := c.MustGet("DB").(*sql.DB)
	siren := c.Param("siren")
	// TODO: valider SIREN
	if siren == "" {
		c.JSON(400, "SIREN obligatoire")
		return
	}
	entreprise := Entreprise{Siren: siren}
	err := entreprise.findEntrepriseBySiren(db)
	if err != nil {
		c.JSON(500, err.Error())
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
	log.Println("getEtablissement")
	db := c.MustGet("DB").(*sql.DB)
	siret := c.Param("siret")
	// TODO: valider SIRET
	if siret == "" {
		c.JSON(400, "SIRET obligatoire")
		return
	}
	etablissement := Etablissement{Siret: siret}
	err := etablissement.findEtablissementBySiret(db)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, etablissement)
}

func getEntrepriseEtablissements(c *gin.Context) {
	log.Println("getEntrepriseEtablissements")
	db := c.MustGet("DB").(*sql.DB)
	siren := c.Param("siren")
	// TODO: valider SIREN
	if siren == "" {
		c.JSON(400, "SIREN obligatoire")
		return
	}
	entreprise := Entreprise{Siren: siren}
	etablissements, err := entreprise.findEtablissementsBySiren(db)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, etablissements)
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
	log.Println("getListes")
	db := c.MustGet("DB").(*sql.DB)
	listes, err := findAllListes(db)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, listes)
}

func getLastListeScores(c *gin.Context) {
	log.Println("getLastListeScores")
	db := c.MustGet("DB").(*sql.DB)
	scores, err := findLastListeScores(db)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, scores)
}

func getListeScores(c *gin.Context) {
	log.Println("getListeScores")
	db := c.MustGet("DB").(*sql.DB)
	listeID := c.Param("listeId")
	if listeID == "" {
		c.JSON(400, "Identifiant de liste obligatoire")
		return
	}
	liste := Liste{ID: listeID}
	scores, err := liste.findListeScores(db)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, scores)
}

func getCodesNaf(c *gin.Context) {
	log.Println("getCodesNaf")
	c.JSON(200, "getCodesNaf")
}

func getDepartements(c *gin.Context) {
	log.Println("getDepartements")
	c.JSON(200, "getDepartements")
}

func getRegions(c *gin.Context) {
	log.Println("getRegions")
	c.JSON(200, "getRegions")
}
