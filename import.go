package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
)

type importObject struct {
	ID    string      `json:"_id"`
	Value importValue `json:"value"`
}

type importValue struct {
	Sirets   *[]string      `json:"sirets"`
	Bdf      *[]interface{} `json:"bdf"`
	Diane    *[]diane       `json:"diane"`
	SireneUL *sireneUL      `json:"sirene_ul"`
}

func processEntreprise(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Buffer([]byte{}, 10000000)

	i := 0
	for scanner.Scan() {
		var e entreprise
		if i > 0 {
			break
		}
		json.Unmarshal(scanner.Bytes(), &e)
		spew.Dump(e)
		i++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func processEtablissement(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Buffer([]byte{}, 1000000)

	i := 0
	for scanner.Scan() {
		var e etablissement
		if i > 10 {
			break
		}
		json.Unmarshal(scanner.Bytes(), &e)
		spew.Dump(e)
		i++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}
