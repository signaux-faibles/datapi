package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func processEntreprise(fileName string, tx *sql.Tx) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Print("error opening file: " + err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Buffer([]byte{}, 10000000)

	i := 0
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}

		if i >= 0 {
			break
		}
		var e entreprise

		json.Unmarshal(scanner.Bytes(), &e)
		if e.Value.SireneUL.Siren != "" {
			err = e.insert(tx)
		}
		if err != nil {
			return err
		}

		i++
	}

	return nil
}

func processEtablissement(fileName string, tx *sql.Tx) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Print("error opening file: " + err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Buffer([]byte{}, 10000000)

	i := 0
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}

		if i <= 4000 {
			var e etablissement
			fmt.Println(string(scanner.Bytes()))
			json.Unmarshal(scanner.Bytes(), &e)
			if len(e.Value.Key) == 14 {
				err = e.insert(tx)
			}
			if err != nil {
				return err
			}

		} else {
			break
		}

		i++
	}

	return nil
}
