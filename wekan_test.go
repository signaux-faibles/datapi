package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/cnf/structhash"
	"github.com/spf13/viper"
)

func loadCardsFromFile(cards *WekanCards, path string) error {
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(fileContent, cards)
	return err
}

func loadFollowExportsFromFile(exports *dbExports, path string) error {
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(fileContent, exports)
	return err
}

func readTestData() (WekanExports, error) {
	var wc WekanConfig
	err := wc.loadFile("test/wekan/wekanConfig.json")
	if err != nil {
		return nil, errors.New("can't read wekan config")
	}

	var cards WekanCards
	err = loadCardsFromFile(&cards, "test/wekan/cards.json")
	if err != nil {
		return nil, errors.New("can't read wekan cards")
	}

	var exports dbExports
	err = loadFollowExportsFromFile(&exports, "test/wekan/followExports.json")
	if err != nil {
		return nil, errors.New("can't read follow exports: " + err.Error())
	}
	return joinExports(wc, exports, cards), nil
}
func Test_WekanExportsDOCX(t *testing.T) {
	t.Log("WekanExports can generate a non-zero length docx file")
	wekanExports, err := readTestData()
	if err != nil {
		t.Error(err.Error())
		return
	}
	viper.Set("docxifyPath", "./docxify3.py")
	viper.Set("docxifyWorkingDir", "./build-container")
	viper.Set("docxifyPython", "python3")
	dateHeader, _ := time.Parse("02/01/2006", "05/06/2018")
	header := ExportHeader{
		Auteur: "test_auteur",
		Date:   dateHeader,
	}
	docx, err := wekanExports.docx(header)
	if len(docx) == 0 {
		t.Error("empty docx file returned")
	}
	if err != nil {
		t.Error(err)
	}
	if os.Getenv("WRITE_DOCX") == "true" {
		err := ioutil.WriteFile("test_output.docx", docx, 0755)
		if err != nil {
			t.Errorf("could create result file: %s", err.Error())
		}
	}
}

func Test_WekanExportsXLSX(t *testing.T) {
	t.Log("WekanExports can generate a non-zero length xlsx file")
	wekanExports, err := readTestData()
	if err != nil {
		t.Error(err.Error())
		return
	}
	viper.Set("docxifyPath", "./docxify3.py")
	viper.Set("docxifyWorkingDir", "./build-container")
	viper.Set("docxifyPython", "python3")
	xlsx, err := wekanExports.xlsx(true)
	if len(xlsx) == 0 {
		t.Error("empty xlsx file returned")
	}
	if err != nil {
		t.Error(err)
	}
}

func Test_WekanExports(t *testing.T) {
	t.Log("joinExports() works as expected")
	wekanExports, err := readTestData()
	if err != nil {
		t.Error(err.Error())
		return
	}
	hash, err := structhash.Hash(wekanExports, 0)
	if err != nil {
		t.Errorf("could not hash WekanExport: %s", err.Error())
	}
	expected := "76305f3733613965313062623830346239323431343737343765343131633034383263"
	if fmt.Sprintf("%x", hash) != expected {
		t.Errorf("unexpected results from joinExports(): \nstructhash should be: %s\nbut structHash is: %x", expected, hash)
	}
}
