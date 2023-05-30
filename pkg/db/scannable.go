package db

import (
	"context"
)

// Scannable est une interface permettant l'usage de la fonction Query
// la fonction Items() doit retourner un slice de pointeurs des valeurs
// recevant une nouvelle ligne de résultat de la requête
type Scannable interface {
	Tuple() []interface{}
}

// Query est une fonction permettant l'exécution d'une requête sql et la récupération des résultats dans un slice
func Query(ctx context.Context, scannable Scannable, sql string, params ...interface{}) error {
	rows, err := db.Query(ctx, sql, params...)
	if err != nil {
		return err
	}
	for rows.Next() {
		items := scannable.Tuple()
		err = rows.Scan(items...)
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
	}
	return nil
}
