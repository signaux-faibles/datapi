package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
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

func readTestData() (Cards, error) {
	var wekanCards WekanCards
	wekanConfig.mu = &sync.Mutex{}
	err := loadCardsFromFile(&wekanCards, "test/wekan/cards.json")
	if err != nil {
		return nil, errors.New("can't read wekan cards")
	}

	var cardIndex = make(map[string]*Card)
	var cards Cards

	var exports dbExports
	err = loadFollowExportsFromFile(&exports, "test/wekan/followExports.json")
	if err != nil {
		return nil, errors.New("can't read follow exports: " + err.Error())
	}

	for _, e := range exports {
		card := Card{
			dbExport: e,
		}
		cardIndex[e.Siret] = &card
		cards = append(cards, &card)
	}

	for _, we := range wekanCards {
		siret, _ := we.Siret()
		if card, ok := cardIndex[siret]; ok {
			card.WekanCards = []*WekanCard{&we}
		}
	}

	return cards, nil
}

func Test_WekanExportsDOCX(t *testing.T) {
	t.Log("WekanExports can generate a non-zero length docx file")
	cards, err := readTestData()

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
	var docxs Docxs
	for _, card := range cards {
		docx, err := card.docx(header)
		if err != nil {
			t.Fail()
		}
		docxs = append(docxs, docx)
	}
	data := docxs.zip()
	if len(data) == 0 {
		t.Error("empty docx file returned")
	}
	if err != nil {
		t.Error(err)
	}
	if os.Getenv("WRITE_DOCX") == "true" {
		err := ioutil.WriteFile("test_output.docx", data, 0755)
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
	expected := "76305f3266663335653439626432636163386564343133323462646339373731346330"
	if fmt.Sprintf("%x", hash) != expected {
		t.Errorf("unexpected results from joinExports(): \nstructhash should be: %s\nbut structHash is: %x", expected, hash)
	}
}
