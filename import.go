package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sync"

	pgx "github.com/jackc/pgx/v4"
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

func processEntreprise(fileName string, tx *pgx.Tx) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("error opening file: %s", err.Error())
		return err
	}
	defer file.Close()
	scanner, err := getFileScanner(file)

	var batches = make(chan *pgx.Batch, 10)
	var wg sync.WaitGroup
	wg.Add(1)
	go runBatches(tx, batches, &wg)
	i := 0
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}

		var e entreprise
		json.Unmarshal(scanner.Bytes(), &e)
		if e.Value.SireneUL.Siren != "" {
			batch := e.getBatch()
			batches <- batch
			i++
			if math.Mod(float64(i), 100) == 0 {
				fmt.Printf("\033[2K\r%s: %d objects inserted", fileName, i)
			}
		}

	}
	close(batches)
	wg.Wait()
	return nil
}

func processEtablissement(fileName string, tx *pgx.Tx) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("error opening file: %s", err.Error())
		return err
	}
	defer file.Close()
	scanner, err := getFileScanner(file)
	if err != nil {
		return err
	}

	var batches = make(chan *pgx.Batch, 10)
	var wg sync.WaitGroup
	wg.Add(1)
	go runBatches(tx, batches, &wg)

	i := 0
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		var e etablissement
		json.Unmarshal(scanner.Bytes(), &e)
		if len(e.ID) > 14 {
			e.Value.Key = e.ID[len(e.ID)-14:]
			batches <- e.getBatch()
			i++
			if math.Mod(float64(i), 100) == 0 {
				fmt.Printf("\033[2K\r%s: %d objects sent to postgres", fileName, i)
			}
		}
	}
	close(batches)
	wg.Wait()
	return nil
}

func getFileScanner(file *os.File) (*bufio.Scanner, error) {
	unzip, err := gzip.NewReader(file)
	if err != nil {
		log.Printf("gzip didn't make it: %s", err.Error())
		return nil, err
	}
	scanner := bufio.NewScanner(unzip)
	scanner.Buffer([]byte{}, 10000000)

	return scanner, nil
}

func upgradeVersion(tx *pgx.Tx) error {
	fmt.Println("Not implemented")
	return nil
}
