package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	rncs "github.com/signaux-faibles/libinpirncs"
	"github.com/spf13/viper"
)

func rncsImportHandler(c *gin.Context) {
	input, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, "rncsImportHandler: "+err.Error())
	}
	bilan, err := rncs.ParseBilan(input, "")
	if err != nil {
		c.JSON(400, "ParseBilan: "+err.Error())
	}
	err = updateDianeFromRncs([]rncs.Bilan{bilan})
	if err != nil {
		c.JSON(500, "updateDianeFromRncs: "+err.Error())
	}
	c.JSON(501, "not implemented")
}

func rncsImportFilesHandler(c *gin.Context) {
	path := viper.GetString("rncsPath")
	var bilanFiles []string
	err := filepath.Walk(path, rncsImportFile(&bilanFiles))
	if err != nil {
		c.JSON(500, "rncsImportFolderHandler: "+err.Error())
	}
	bilans, err := getBilansFromFiles(bilanFiles)
	if err != nil {
		c.JSON(500, "getBilanFromFiles: "+err.Error())
	}
	fmt.Println(bilans)
}

func rncsImportFile(bilansFiles *[]string) func(string, os.FileInfo, error) error {
	return func(path string, file os.FileInfo, err error) error {
		l := len(file.Name())
		if len(file.Name()) > 3 && strings.ToLower(file.Name()[l-3:l]) == "xml" {
			*bilansFiles = append(*bilansFiles, path)
		}
		return nil
	}
}

func getBilansFromFiles(bilanFiles []string) ([]rncs.Bilan, error) {
	var bilans []rncs.Bilan
	for _, file := range bilanFiles {
		if bilanData, err := ioutil.ReadFile(file); err == nil {
			if bilan, err := rncs.ParseBilan(bilanData, file); err == nil {
				bilans = append(bilans, bilan)
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return bilans, nil
}

func updateDianeFromRncs(bilans []rncs.Bilan) error {
	for _, b := range bilans {
		d, err := rncsToDiane(b)
	}
	return nil
}

func rncsToDiane(rncs.Bilan) (diane, error) {

	return diane{}, nil
}
