package core

import (
	"context"
	"datapi/pkg/db"
)

func loadDepartementReferentiel() (map[CodeDepartement]string, error) {
	rows, err := db.Get().Query(context.Background(), "select code, libelle from departements")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	r := make(map[CodeDepartement]string)
	for rows.Next() {
		var code CodeDepartement
		var libelle string
		err := rows.Scan(&code, &libelle)
		if err != nil {
			return nil, err
		}
		r[code] = libelle
	}
	return r, nil
}

func loadRegionsReferentiel() (map[Region][]CodeDepartement, error) {
	rows, err := db.Get().Query(context.Background(), `
		select r.libelle, d.code from regions r
		inner join departements d on d.id_region = r.id
    	union select 'France entière' as libelle, code from departements
		order by libelle, code
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	r := make(map[Region][]CodeDepartement)
	for rows.Next() {
		var region Region
		var departement CodeDepartement
		err := rows.Scan(&region, &departement)
		if err != nil {
			return nil, err
		}
		r[region] = append(r[region], departement)
	}
	return r, nil
}
