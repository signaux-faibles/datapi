package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
	"regexp"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// BilansInpi porte les metadonnées des bilans dans la réponse de l'API entreprise
type BilansInpi struct {
	URLDocuments string `json:"url_documents"`
	Bilans       []struct {
		IDFichier           int         `json:"id_fichier"`
		Siren               string      `json:"siren"`
		DenominationSociale interface{} `json:"denomination_sociale"`
		CodeGreffe          int         `json:"code_greffe"`
		DateDepot           string      `json:"date_depot"`
		NatureArchive       string      `json:"nature_archive"`
		Confidentiel        int         `json:"confidentiel"`
		DateCloture         time.Time   `json:"date_cloture"`
		NumeroGestion       string      `json:"numero_gestion"`
	} `json:"bilans"`
}

func bilansExercicesHandler(c *gin.Context) {
	siren := c.Param("siren")
	userID := c.GetString("userID")
	body, err := makeBilansInpiRequest(userID, siren)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	bilansInpi := BilansInpi{}
	bodyString := string(body)
	err = json.Unmarshal([]byte(bodyString), &bilansInpi)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	exercices := make([]int, 0)
	for _, bilan := range bilansInpi.Bilans {
		exercice, err := strconv.Atoi(bilan.DateCloture.String()[:4])
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		exercices = append(exercices, exercice)
	}
	sort.Ints(exercices)
	c.JSON(200, exercices)
}

func bilansDocumentsHandler(c *gin.Context) {
	siren := c.Param("siren")
	exercice := c.Param("exercice")
	match, err := regexp.MatchString("^[0-9]{4}$", exercice)
	if err != nil || !match {
		t := time.Now()
		lastExercice := t.Year() - 1
		exercice = strconv.Itoa(lastExercice)
	}
	ok := false
	data := make([]byte, 0)
	key := "documents_comptables-" + siren + "-" + exercice
	cached, inCache := getCachedFile(key)
	if inCache {
		data = cached
		ok = true
	} else {
		userID := c.GetString("userID")
		body, err := makeBilansInpiRequest(userID, siren)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		bilansInpi := BilansInpi{}
		bodyString := string(body)
		err = json.Unmarshal([]byte(bodyString), &bilansInpi)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		resp, err := http.Get(bilansInpi.URLDocuments)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		for _, parentFile := range zipReader.File {
			if strings.HasSuffix(parentFile.Name, "zip") {
				parts := strings.Split(parentFile.Name, "_")
				if parts[0] == "CA" && parts[4] == exercice {
					unzippedFileBytes, err := readZipFile(parentFile)
					if err != nil {
						c.JSON(500, err.Error())
						return
					}
					zipReader, err := zip.NewReader(bytes.NewReader(unzippedFileBytes), int64(len(unzippedFileBytes)))
					if err != nil {
						c.JSON(500, err.Error())
						return
					}
					for _, childFile := range zipReader.File {
						if strings.HasSuffix(childFile.Name, "pdf") {
							unzippedFileBytes, err := readZipFile(childFile)
							if err != nil {
								c.JSON(500, err.Error())
								return
							}
							data = unzippedFileBytes
							ok = true
							setCachedFile(key, data)
						}
					}
				}
			}
		}
	}
	if ok {
		c.Header("Content-Disposition", "attachment; filename="+key+".pdf")
		c.Data(200, "application/octet-stream", data)
	} else {
		c.JSON(204, "Bilan indisponible")
	}
}

func makeBilansInpiRequest(userID string, siren string) ([]byte, error) {
	// doc: https://entreprise.api.gouv.fr/catalogue/#a-bilans_inpi
	url := viper.GetString("apiEntrepriseBilansInpiUrl")
	req, err := http.NewRequest("GET", url + siren, nil)
	if err != nil {
		return nil, err
	}
	token := viper.GetString("apiEntrepriseToken")
	req.Header.Set("Authorization", "Bearer " + token)
	q := req.URL.Query()
	context := viper.GetString("apiEntrepriseContext")
	q.Add("context", context)
	recipient := viper.GetString("apiEntrepriseRecipient")
	q.Add("recipient", recipient)
	object := viper.GetString("apiEntrepriseObject") + userID
	q.Add("object", object)
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func buildFilename(key string) string {
	cacheDirectory := viper.GetString("fileCachePath")
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return path.Join(cacheDirectory, hex.EncodeToString(hasher.Sum(nil)))
}

func setCachedFile(key string, data []byte) error {
	filename := buildFilename(key)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func getCachedFile(key string) ([]byte, bool) {
	filename := buildFilename(key)
	f, err := os.Open(filename)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, false
	}
	return data, true
}