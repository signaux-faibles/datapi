package main

import (
	"context"
	"fmt"
	"testing"
)

func TestListes(t *testing.T) {
	_, indented, _ := get(t, "/listes")
	diff, _ := processGoldenFile(t, "data/listes.json.gz", indented)
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: \n%s", diff)
	}
}

func TestScores(t *testing.T) {
	t.Log("/scores/liste retourne le même résultat qu'attendu")
	_, indented, _ := post(t, "/scores/liste", nil)
	diff, _ := processGoldenFile(t, "data/scores.json.gz", indented)
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: \n%s", diff)
	}

	t.Log("/scores/liste retourne le même résultat qu'attendu avec ignoreZone=true")
	params := map[string]interface{}{
		"ignoreZone": true,
	}
	_, indented, _ = post(t, "/scores/liste", params)
	diff, _ = processGoldenFile(t, "data/scores-ignoreZone.json.gz", indented)
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: \n%s", diff)
	}
}

func TestSearch(t *testing.T) {
	t.Log("/etablissement/search retourne 400")
	resp, _, _ := get(t, "/etablissement/search/t")
	if resp.StatusCode != 400 {
		t.Errorf("mauvais status retourné: %d", resp.StatusCode)
	}

}

func TestGetEtablissement(t *testing.T) {
	// récupérer une liste de sirets à chercher
	rows, err := db.Query(context.Background(), `select e.siret from etablissement e
	inner join departements d on d.code = e.departement
	inner join regions r on r.id = d.id_region
	where r.libelle in ('Bourgogne-Franche-Comté', 'Auvergne-Rhône-Alpes')
	order by e.siret
	limit 10`)
	if err != nil {
		t.Errorf("impossible de se connecter à la base: %s", err.Error())
	}

	i := 0
	for rows.Next() {
		var siret string
		err := rows.Scan(&siret)
		if err != nil {
			t.Errorf("siret illisible: %s", err.Error())
		}

		t.Logf("l'établissement %s est bien de la forme attendue", siret)
		_, indented, _ := get(t, "/etablissement/get/"+siret)
		goldenFilePath := fmt.Sprintf("data/etablissement-%d.json.gz", i)
		diff, _ := processGoldenFile(t, goldenFilePath, indented)
		if diff != "" {
			t.Errorf("differences entre le résultat et le golden file: 'data/etablissement-%d.json.gz' \n%s", i, diff)
		}
		i++
	}
}
