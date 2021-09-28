package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// WekanConfig type pour le fichier de configuration de Wekan
type WekanConfig struct {
	Boards map[string]struct {
		BoardID      string            `json:"boardId"`
		Slug         string            `json:"slug"`
		Swimlanes    map[string]string `json:"swimlanes"`
		Lists        []string          `json:"lists"`
		CustomFields struct {
			SiretField    string `json:"siretField"`
			ActiviteField string `json:"activiteField"`
			EffectifField struct {
				EffectifFieldID    string   `json:"effectifFieldId"`
				EffectifFieldItems []string `json:"effectifFieldItems"`
			} `json:"effectifField"`
			FicheSFField string `json:"ficheSFField"`
		} `json:"customFields"`
	} `json:"boards"`
	Users map[string]string `json:"users"`
}

// func (wc WekanConfig) boards() map[string]string {
// 	boards := make(map[string]string)
// 	for r, b := range wc.Boards {
// 		boards[b.BoardID] = r
// 	}
// 	return boards
// }

func (wc WekanConfig) siretFields() []string {
	var siretFields []string
	for _, b := range wc.Boards {
		siretFields = append(siretFields, b.CustomFields.SiretField)
	}
	return siretFields
}

func (wc *WekanConfig) load() error {
	configFile := viper.GetString("wekanConfigFile")
	return wc.loadFile(configFile)
}

func (wc *WekanConfig) loadFile(path string) error {
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(fileContent, wc)
	return err
}

// type wekanDbCustomFields struct {
// 	ID    string `json:"_id" bson:"_id"`
// 	Value string `json:"value" bson:"value"`
// }

// WekanCards est une liste de WekanCard
type WekanCards []WekanCard

// WekanCard embarque la sélection de champs Wekan nécessaires au traitement d'export
type WekanCard struct {
	ID          string    `json:"_id" bson:"_id"`
	Description string    `json:"description" bson:"description"`
	StartAt     time.Time `json:"startAt" bson:"startAt"`
	Siret       string    `json:"siret" bson:"siret"`
}

// Token type pour faire persister en mémoire les tokens réutilisables
type Token struct {
	Value        string
	CreationDate time.Time
}

// WekanLogin type pour les réponses du service de login
type WekanLogin struct {
	ID           string `json:"id"`
	Token        string `json:"token"`
	TokenExpires string `json:"tokenExpires"`
}

// WekanCreateToken type pour les réponses du service de création de token
type WekanCreateToken struct {
	ID        string `json:"_id"`
	AuthToken string `json:"authToken"`
}

// WekanGetCard type pour les réponses du service de recherche de carte
type WekanGetCard []struct {
	ID          string `json:"_id"`
	ListID      string `json:"listId"`
	Description string `json:"description"`
}

// WekanCreateCard type pour les réponses du service de création de carte
type WekanCreateCard struct {
	ID string `json:"_id"`
}

// CustomField type pour les champs personnalisés lors de l'édition de carte
type CustomField struct {
	ID    string `json:"_id"`
	Value string `json:"value"`
}

// FicheCRP type pour les fiches CRP
type FicheCRP struct {
	Siret       string
	Commentaire string
	Description string
}

// EtablissementData type pour les données de l'établissement à faire figurer sur la carte Wekan
type EtablissementData struct {
	Siret           string
	RaisonSociale   string
	Departement     string
	Region          string
	Effectif        int
	CodeActivite    string
	LibelleActivite string
}

var tokens sync.Map

//TODO: factorisation
func wekanGetCardHandler(c *gin.Context) {
	siret := c.Param("siret")
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
		return
	}
	username := c.GetString("username")
	userID, ok := config.Users[username]
	if !ok {
		c.JSON(403, "not a wekan user")
		return
	}
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
	etsData, err := getEtablissementDataFromDb(siret)
	if err != nil {
		log.Println(err.Error())
		return
	}
	region := etsData.Region
	board, ok := config.Boards[region]
	if !ok {
		c.JSON(500, "missing board in config file")
		return
	}
	boardID := board.BoardID
	siretField := board.CustomFields.SiretField
	body, err := getCard(userToken.Value, boardID, siretField, siret)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var wekanGetCard WekanGetCard
	err = json.Unmarshal(body, &wekanGetCard)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	if len(wekanGetCard) == 0 {
		c.JSON(404, "card not found")
		return
	}
	card := wekanGetCard[0]
	cardID := card.ID
	cardDescription := card.Description
	lists := board.Lists
	listIndex := indexOf(card.ListID, lists)
	wekanURL := viper.GetString("wekanURL")
	cardURL := wekanURL + "b/" + boardID + "/" + board.Slug + "/" + cardID
	cardData := map[string]interface{}{
		"cardId":          cardID,
		"listIndex":       listIndex,
		"cardURL":         cardURL,
		"cardDescription": cardDescription,
	}
	c.JSON(200, cardData)
}

func wekanNewCardHandler(c *gin.Context) {
	siret := c.Param("siret")
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
		return
	}
	username := c.GetString("username")
	userID, ok := config.Users[username]
	if !ok {
		c.JSON(403, "not a wekan user")
		return
	}
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
	// data enrichment
	etsData, err := getEtablissementDataFromDb(siret)
	if err != nil {
		log.Println(err.Error())
		return
	}
	region := etsData.Region
	board, ok := config.Boards[region]
	if !ok {
		c.JSON(500, "missing board in config file")
		return
	}
	boardID := board.BoardID
	listID := board.Lists[0]
	// fields formatting
	departement := etsData.Departement
	swimlaneID, ok := board.Swimlanes[departement]
	if !ok {
		c.JSON(500, "not a departement of this region")
		return
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
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var requestData map[string]string
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	description := requestData["description"]
	if description != "" {
		creationData["description"] = description
	}
	body, err = createCard(userToken.Value, boardID, listID, creationData)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	wekanCreateCard := WekanCreateCard{}
	err = json.Unmarshal(body, &wekanCreateCard)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	cardID := wekanCreateCard.ID
	// card edition
	var customFields = []CustomField{
		{board.CustomFields.SiretField, siret},
		{board.CustomFields.FicheSFField, ficheSF},
	}
	if activite != "" {
		customFields = append(customFields, CustomField{board.CustomFields.ActiviteField, activite})
	}
	if effectifIndex > 0 {
		customFields = append(customFields, CustomField{board.CustomFields.EffectifField.EffectifFieldID, board.CustomFields.EffectifField.EffectifFieldItems[effectifIndex]})
	}
	now := time.Now()
	var editionData = map[string]interface{}{
		"customFields": customFields,
		"startAt":      now,
	}
	_, err = editCard(adminToken.Value, boardID, listID, cardID, editionData)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
}

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
		}
		if effectifIndex > 0 {
			customFields = append(customFields, CustomField{board.CustomFields.EffectifField.EffectifFieldID, board.CustomFields.EffectifField.EffectifFieldItems[effectifIndex]})
		} else {
			log.Printf("unknown effectif class")
		}
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

func wekanGetListCardsHandler(c *gin.Context) {
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
		return
	}
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
	region := "Auvergne-Rhône-Alpes"
	board, ok := config.Boards[region]
	if !ok {
		c.JSON(500, "missing board in config file")
		return
	}
	boardID := board.BoardID
	var cardDataArray = []map[string]interface{}{}
	for _, listID := range board.Lists[:4] {
		log.Printf("listID=%s", listID)
		body, err := getListCards(adminToken.Value, boardID, listID)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		var wekanGetCard WekanGetCard
		err = json.Unmarshal(body, &wekanGetCard)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		if len(wekanGetCard) == 0 {
			c.JSON(404, "card not found")
			return
		}
		for _, card := range wekanGetCard {
			cardID := card.ID
			cardDescription := card.Description
			lists := board.Lists
			listIndex := indexOf(card.ListID, lists)
			wekanURL := viper.GetString("wekanURL")
			cardURL := wekanURL + "b/" + boardID + "/" + board.Slug + "/" + cardID
			cardData := map[string]interface{}{
				"cardId":          cardID,
				"listIndex":       listIndex,
				"cardURL":         cardURL,
				"cardDescription": cardDescription,
			}
			cardDataArray = append(cardDataArray, cardData)
		}
	}
	log.Printf("cardDataArray=%d", len(cardDataArray))
	c.JSON(200, "OK")
}

func adminLogin(username string, password string) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	req, err := http.NewRequest("POST", wekanURL+"users/login", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func createToken(userID string, adminToken string) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	req, err := http.NewRequest("POST", wekanURL+"api/createtoken/"+userID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+adminToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getListCards(userToken string, boardID string, listID string) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	req, err := http.NewRequest("GET", wekanURL+"api/boards/"+boardID+"/lists/"+listID+"/cards", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+userToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		err = errors.New("unexpected wekan API response")
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getCard(userToken string, boardID string, siretField string, siret string) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	req, err := http.NewRequest("GET", wekanURL+"api/boards/"+boardID+"/cardsByCustomField/"+siretField+"/"+siret, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+userToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		err = errors.New("unexpected wekan API response")
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func createCard(userToken string, boardID string, listID string, creationData map[string]interface{}) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	json, err := json.Marshal(creationData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", wekanURL+"api/boards/"+boardID+"/lists/"+listID+"/cards", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+userToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		err = errors.New("unexpected wekan API response")
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func editCard(userToken string, boardID string, listID string, cardID string, editionData map[string]interface{}) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	json, err := json.Marshal(editionData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PUT", wekanURL+"api/boards/"+boardID+"/lists/"+listID+"/cards/"+cardID, bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+userToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		err = errors.New("unexpected wekan API response")
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func commentCard(userToken string, boardID string, cardID string, commentData map[string]string) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	json, err := json.Marshal(commentData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", wekanURL+"api/boards/"+boardID+"/cards/"+cardID+"/comments", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+userToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		err = errors.New("unexpected wekan API response")
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func isValidToken(token Token) bool {
	duration := time.Since(token.CreationDate)
	hours := duration.Hours()
	days := hours / 24
	return days < float64(viper.GetInt("wekanTokenDays"))
}

func indexOf(element string, array []string) int {
	for i, v := range array {
		if element == v {
			return i
		}
	}
	return -1
}

func getEtablissementDataFromDb(siret string) (EtablissementData, error) {
	sql := `select s.siret,
	coalesce(s.raison_sociale, ''), 
	coalesce(s.code_departement, ''), 
	coalesce(r.libelle, ''),
	coalesce(s.effectif, 0),
	coalesce(s.code_activite, ''), 
	coalesce(s.libelle_n5, '') 
	from v_summaries s
	inner join departements d on d.code = s.code_departement
	inner join regions r on r.id = d.id_region
	where s.siret = '$1`
	rows, err := db.Query(context.Background(), sql, siret)
	var etsData EtablissementData
	if err != nil {
		return etsData, err
	}
	for rows.Next() {
		err := rows.Scan(&etsData.Siret, &etsData.RaisonSociale, &etsData.Departement, &etsData.Region, &etsData.Effectif, &etsData.CodeActivite, &etsData.LibelleActivite)
		if err != nil {
			return etsData, err
		}
	}
	return etsData, nil
}

func formatActiviteField(codeActivite string, libelleActivite string) string {
	var activite string
	if len(libelleActivite) > 0 && len(codeActivite) > 0 {
		activite = libelleActivite + " (" + codeActivite + ")"
	}
	return activite
}

func getEffectifIndex(effectif int) int {
	var effectifIndex = -1
	effectifClass := [4]int{10, 20, 50, 100}
	for i := len(effectifClass) - 1; i >= 0; i-- {
		if effectif >= effectifClass[i] {
			effectifIndex = i
			break
		}
	}
	return effectifIndex
}

func formatFicheSFField(siret string) string {
	return viper.GetString("webBaseURL") + "ets/" + siret
}
