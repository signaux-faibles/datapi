package wekan

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"datapi/pkg/core"
)

func getEtablissementDataFromDb(ctx context.Context, db *pgxpool.Pool, siret core.Siret) (core.EtablissementData, error) {
	sql := `select s.siret,
	coalesce(s.raison_sociale, ''),
	coalesce(s.code_departement, ''),
	coalesce(r.libelle, ''),
	coalesce(s.effectif, 0),
	coalesce(s.code_activite, ''),
	coalesce(s.libelle_n5, '')
	from v_summaries s
	inner join departements d on d.code = s.code_departement
	inner join regions r on r.id = d.id_region
	where s.siret = $1`
	rows, err := db.Query(ctx, sql, siret)
	var etsData core.EtablissementData
	if err != nil {
		return etsData, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&etsData.Siret, &etsData.RaisonSociale, &etsData.Departement, &etsData.Region, &etsData.Effectif, &etsData.CodeActivite, &etsData.LibelleActivite)
		if err != nil {
			return etsData, err
		}
	}
	return etsData, nil
}
