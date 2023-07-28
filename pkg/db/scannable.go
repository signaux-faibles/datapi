package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Scannable est une interface permettant l'usage de la fonction Scan
// la fonction Items() doit retourner un slice de pointeurs des valeurs
// recevant une nouvelle ligne de résultat de la requête
type Scannable interface {
	Tuple() []interface{}
}

// Scan est une fonction permettant l'exécution d'une requête sql et la récupération des résultats dans un slice
func Scan(ctx context.Context, scannable Scannable, sql string, params ...interface{}) error {
	return SelectTuples(ctx, Get(), scannable, sql, params)
}

// SelectTuples est une fonction permettant l'exécution d'une requête sql et la récupération des résultats dans un slice
func SelectTuples(ctx context.Context, dbPool *pgxpool.Pool, scannable Scannable, sql string, params ...interface{}) error {
	rows, err := dbPool.Query(ctx, sql, params...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		tuple := scannable.Tuple()
		err = rows.Scan(tuple...)
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
	}
	return nil
}
