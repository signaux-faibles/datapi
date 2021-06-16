package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
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

func Test_ExportWekan(t *testing.T) {
	var wc WekanConfig
	err := wc.loadFile("test/wekan/wekanConfig.json")
	if err != nil {
		t.Error("can't read wekan config")
	}

	var cards WekanCards
	err = loadCardsFromFile(&cards, "test/wekan/cards.json")
	if err != nil {
		t.Error("can't read wekan cards")
	}

	var exports dbExports
	err = loadFollowExportsFromFile(&exports, "test/wekan/followExports.json")
	if err != nil {
		t.Error("can't read follow exports: " + err.Error())
	}

	wekanExports := joinExports(wc, exports, cards)
	file, err := wekanExports.docx()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(file))
}
