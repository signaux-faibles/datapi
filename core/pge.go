package core

import (
	"github.com/jackc/pgx/v4"
)

func (e *Etablissements) addPGEsSelection(batch *pgx.Batch, roles scope, username string) {
	batch.Queue(
		`select 
					e.siren,
       				bool_or(e.actif) actif
				from 
					entreprise_pge e
					inner join f_etablissement_permissions($1, $2) p on p.siren = e.siren and pge
				where 
					e.siren=any($3)
				group by e.siren
				order by e.siren;`,
		roles.zoneGeo(), username, e.sirensFromQuery())
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
