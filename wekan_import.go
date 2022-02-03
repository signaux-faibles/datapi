package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func wekanImportHandler(c *gin.Context) {
	region := c.Query("region")
	if region == "" {
		c.JSON(400, "region query string missing")
		return
	}
	username := c.Query("username")
	if username == "" {
		c.JSON(400, "username query string missing")
		return
	}
	log.Printf("1. csv file parsing")
	importFile := viper.GetString("wekanImportFile")
	csvFile, err := os.Open(importFile)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer csvFile.Close()
	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		log.Println(err.Error())
		return
	}
	var tableauCRP []FicheCRP
	var invalidLines []string
	reSiret, _ := regexp.Compile("^[0-9]{14}$")
	for i, line := range csvLines {
		if i > 1 {
			siret := strings.ReplaceAll(line[1], " ", "")
			match := reSiret.MatchString(siret)
			if !match {
				invalidLines = append(invalidLines, "#"+strconv.Itoa(i+1))
			} else {
				commentaire := "**Difficultés rencontrées par l'entreprise**\n" + line[16] + "\n\n**Etat du climat social**\n" + line[19] + "\n\n**Etat du dossier**\n" + line[22] + "\n\n**Actions et échéances à venir**\n" + line[24] + "\n\n**Commentaires sur l'entreprise**\n" + line[25] + "\n\n**Commentaires sur la situation**\n" + line[26] + "\n\n"
				description := line[27]
				ficheCRP := FicheCRP{siret, commentaire, description}
				tableauCRP = append(tableauCRP, ficheCRP)
			}
		}
	}
	log.Printf("%d lines to import", len(tableauCRP))
	if len(invalidLines) > 0 {
		log.Printf("%d invalid lines : %s", len(invalidLines), strings.Join(invalidLines, ", "))
	} else {
		log.Printf("no invalid lines")
	}
	log.Printf("2. config file reading")
	configFile := viper.GetString("wekanConfigFile")
	fileContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var config WekanConfig
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		c.AbortWithError(500, err)
	}
	board, ok := config.Boards[region]
	if !ok {
		c.JSON(500, "missing board in config file")
		return
	}
	boardID := board.BoardID
	listID := board.Lists[0]
	userID, ok := config.Users[username]
	if !ok {
		c.JSON(500, "not a wekan user")
		return
	}
	log.Printf("3. admin login")
	var adminToken Token
	adminUsername := viper.GetString("wekanAdminUsername")
	loadedToken, ok := tokens.Load(adminUsername)
	if loadedToken != nil {
		adminToken = loadedToken.(Token)
		ok = isValidToken(adminToken)
	}
	if !ok {
		log.Printf("no valid admin token")
		adminPassword := viper.GetString("wekanAdminPassword")
		body, err := adminLogin(adminUsername, adminPassword)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		wekanLogin := WekanLogin{}
		err = json.Unmarshal(body, &wekanLogin)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		adminToken.Value = wekanLogin.Token
		now := time.Now()
		adminToken.CreationDate = now
		tokens.Store(adminUsername, adminToken)
	} else {
		log.Printf("existing admin token")
	}
	log.Printf("4. user token creation")
	var userToken Token
	loadedToken, ok = tokens.Load(username)
	if loadedToken != nil {
		userToken = loadedToken.(Token)
		ok = isValidToken(userToken)
	}
	if !ok {
		log.Printf("no valid user token")
		body, err := createToken(userID, adminToken.Value)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		wekanCreateToken := WekanCreateToken{}
		err = json.Unmarshal(body, &wekanCreateToken)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		userToken.Value = wekanCreateToken.AuthToken
		now := time.Now()
		userToken.CreationDate = now
		tokens.Store(username, userToken)
	} else {
		log.Printf("existing user token")
	}
	log.Printf("5. wekan cards creation")
	for _, ficheCRP := range tableauCRP {
		siret := ficheCRP.Siret
		log.Printf("ets %s to import...", siret)
		// data enrichment
		etsData, err := getEtablissementDataFromDb(siret)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		// fields formatting
		if siret != etsData.Siret {
			log.Printf("unknown ets")
			continue
		}
		if region != etsData.Region {
			log.Printf("incorrect region")
			continue
		}
		departement := etsData.Departement
		if departement == "" {
			log.Printf("unknown departement")
			continue
		}
		swimlaneID, ok := board.Swimlanes[departement]
		if !ok {
			log.Printf("not a departement of this region")
			continue
		}
		activite := formatActiviteField(etsData.CodeActivite, etsData.LibelleActivite)
		ficheSF := formatFicheSFField(etsData.Siret)
		effectifIndex := getEffectifIndex(etsData.Effectif)
		// new card
		title := etsData.RaisonSociale
		creationData := map[string]interface{}{
			"authorId":   userID,
			"members":    [1]string{userID},
			"title":      title,
			"swimlaneId": swimlaneID,
		}
		description := ficheCRP.Description
		if description != "" {
			creationData["description"] = description
		}
		body, err := createCard(userToken.Value, boardID, listID, creationData)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		wekanCreateCard := WekanCreateCard{}
		err = json.Unmarshal(body, &wekanCreateCard)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		cardID := wekanCreateCard.ID
		// card edition
		var customFields = []CustomField{
			{board.CustomFields.SiretField, siret},
			{board.CustomFields.FicheSFField, ficheSF},
		}
		if activite != "" {
			customFields = append(customFields, CustomField{board.CustomFields.ActiviteField, activite})
		} else {
			log.Printf("unknown activite")
			customFields = append(customFields, CustomField{board.CustomFields.ActiviteField, ""})
		}
		if effectifIndex >= 0 {
			customFields = append(customFields, CustomField{board.CustomFields.EffectifField.EffectifFieldID, board.CustomFields.EffectifField.EffectifFieldItems[effectifIndex]})
		} else {
			log.Printf("unknown effectif class")
			customFields = append(customFields, CustomField{board.CustomFields.EffectifField.EffectifFieldID, ""})
		}
		customFields = append(customFields, CustomField{board.CustomFields.ContactField, ""})
		now := time.Now()
		var editionData = map[string]interface{}{
			"customFields": customFields,
			"startAt":      now,
		}
		_, err = editCard(userToken.Value, boardID, listID, cardID, editionData)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		// card comment
		comment := ficheCRP.Commentaire
		commentData := map[string]string{
			"authorId": userID,
			"comment":  comment,
		}
		_, err = commentCard(userToken.Value, boardID, cardID, commentData)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		log.Printf("ets %s imported", siret)
	}
}
