package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

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
	docx, err := wekanExports.docx()
	if len(docx) == 0 {
		t.Error("empty docx file returned")
	}
	if err != nil {
		t.Error(err)
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
	expected := "76305f3466653530373036626362616331383366646236646366616438623965353663"
	if fmt.Sprintf("%x", hash) != expected {
		t.Errorf("unexpected results from joinExports(): \nstructhash should be: %s\nbut structHash is: %x", expected, hash)
	}
}
