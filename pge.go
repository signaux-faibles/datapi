package main

import (
	"github.com/jackc/pgx/v4"
)

func (e *Etablissements) addPGEsSelection(batch *pgx.Batch) {
	batch.Queue(
		`select
					siren,
					actif
				from 
					entreprise_pge
				where 
					siren=any($1)
				order by siren;`,
		e.sirensFromQuery())
}

func (e Etablissements) loadPGE(rows *pgx.Rows) error {
	var pges = make(map[string][]bool)

	for (*rows).Next() {
		var pgeActif bool
		var siren string
		err := (*rows).Scan(
			&siren,
			&pgeActif,
		)
		if err != nil {
			return err
		}
		pges[siren] = append(pges[siren], pgeActif)
	}
	for siren, pge := range pges {

		entreprise := e.Entreprises[siren]

		var pgeActif = false
		for _, current := range pge {
			pgeActif = pgeActif || current
		}
		entreprise.PGEActif = &pgeActif
		e.Entreprises[siren] = entreprise
	}
	return nil
}
