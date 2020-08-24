package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"

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

	batches, wg := newBatchRunner(tx)
	unzip, err := gzip.NewReader(file)
	decoder := json.NewDecoder(unzip)

	i := 0
	for {
		var e entreprise
		err := decoder.Decode(&e)
		if err != nil {
			close(batches)
			wg.Wait()
			fmt.Println("wait")
			if err == io.EOF {
				fmt.Println("exite")
				return nil
			}
			return err
		}

		if e.Value.SireneUL.Siren != "" {
			batch := e.getBatch()
			fmt.Print("coucou1\n")
			batches <- batch
			fmt.Print("coucou2\n")
			i++
			if math.Mod(float64(i), 100) == 0 {
				fmt.Printf("\033[2K\r%s: %d objects inserted", fileName, i)
			}
		}
	}

}

func processEtablissement(fileName string, tx *pgx.Tx) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("error opening file: %s", err.Error())
		return err
	}
	defer file.Close()
	if err != nil {
		return err
	}

	batches, wg := newBatchRunner(tx)
	unzip, err := gzip.NewReader(file)
	decoder := json.NewDecoder(unzip)

	i := 0
	for {
		var e etablissement
		decoder.Decode(&e)
		if err != nil {
			close(batches)
			wg.Wait()
			if err == io.EOF {
				return nil
			}
			return err
		}

		if len(e.ID) > 14 {
			e.Value.Key = e.ID[len(e.ID)-14:]
			batches <- e.getBatch()
			i++
			if math.Mod(float64(i), 100) == 0 {
				fmt.Printf("\033[2K\r%s: %d objects sent to postgres", fileName, i)
			}
		}
	}

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
