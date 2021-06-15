package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

func loadCardsFromFile(cards *dbWekanCards, path string) error {
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(fileContent, cards)
	return err
}

func loadFollowExportsFromFile(exports *dbFollowExports, path string) error {
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

	var cards dbWekanCards
	err = loadCardsFromFile(&cards, "test/wekan/cards.json")
	if err != nil {
		t.Error("can't read wekan cards")
	}

	var exports dbFollowExports
	err = loadFollowExportsFromFile(&exports, "test/wekan/followExports.json")
	if err != nil {
		t.Error("can't read follow exports: " + err.Error())
	}

	s, _ := json.Marshal(joinExports(wc, exports, cards))
	fmt.Println(string(s))
}
