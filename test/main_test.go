package main

import (
	"context"
	"fmt"
	"testing"
)

func TestListes(t *testing.T) {
	_, indented, _ := get(t, "/listes")
	processGoldenFile(t, "data/listes.json.gz", indented)
}

func TestSearch(t *testing.T) {
	t.Log("/etablissement/search retourne 400")
	resp, _, _ := get(t, "/etablissement/search/t")
	if resp.StatusCode != 400 {
		t.Errorf("mauvais status retourné: %d", resp.StatusCode)
	}

	rows, err := db.Query(context.Background(), `select distinct substring(e.siret from 1 for 3) from etablissement e
	inner join departements d on d.code = e.departement
	inner join regions r on r.id = d.id_region
	where r.libelle in ('Bourgogne-Franche-Comté', 'Auvergne-Rhône-Alpes')
	order by substring(e.siret from 1 for 3)
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

		t.Logf("la recherche %s est bien de la forme attendue", siret)
		_, indented, _ := get(t, "/etablissement/search/"+siret)
		goldenFilePath := fmt.Sprintf("data/search-%d.json.gz", i)
		processGoldenFile(t, goldenFilePath, indented)
		i++
	}
}

func TestFollow(t *testing.T) {
	// récupérer une liste de sirets à suivre de toutes les typologies d'établissements
	sirets := getSiret(t, VIAF{false, false, false, false}, 4)
	sirets = append(sirets, getSiret(t, VIAF{false, false, true, false}, 1)...)
	sirets = append(sirets, getSiret(t, VIAF{false, true, false, false}, 1)...)
	sirets = append(sirets, getSiret(t, VIAF{false, true, true, false}, 1)...)
	sirets = append(sirets, getSiret(t, VIAF{true, false, false, false}, 1)...)
	sirets = append(sirets, getSiret(t, VIAF{true, false, true, false}, 1)...)
	sirets = append(sirets, getSiret(t, VIAF{true, true, false, false}, 1)...)
	sirets = append(sirets, getSiret(t, VIAF{true, true, true, false}, 1)...)

	params := map[string]interface{}{
		"comment":  "test",
		"category": "test",
	}

	for _, siret := range sirets {
		t.Logf("suivi de l'établissement %s", siret)
		resp, _, _ := post(t, "/follow/"+siret, params)
		if resp.StatusCode != 201 {
			t.Errorf("le suivi a échoué: %d", resp.StatusCode)
		}
	}

	for _, siret := range sirets {
		t.Logf("suivi doublon de l'établissement %s", siret)
		resp, _, _ := post(t, "/follow/"+siret, params)
		if resp.StatusCode != 204 {
			t.Errorf("le doublon n'a pas été détecté correctement: %d", resp.StatusCode)
		}
	}

	db.Exec(context.Background(), "update etablissement_follow set since='2020-03-01'")
	_, indented, _ := get(t, "/follow")
	processGoldenFile(t, "data/follow.json.gz", indented)
}

func TestScores(t *testing.T) {
	t.Log("/scores/liste retourne le même résultat qu'attendu")
	_, indented, _ := post(t, "/scores/liste", nil)
	processGoldenFile(t, "data/scores.json.gz", indented)

	t.Log("/scores/liste retourne le même résultat qu'attendu avec ignoreZone=true")
	params := map[string]interface{}{
		"ignoreZone": true,
	}
	_, indented, _ = post(t, "/scores/liste", params)
	processGoldenFile(t, "data/scores-ignoreZone.json.gz", indented)

	t.Log("/scores/liste retourne le même résultat qu'attendu avec ignoreZone=true et exclureSuivi=true")
	params["exclureSuivi"] = true
	_, indented, _ = post(t, "/scores/liste", params)
	processGoldenFile(t, "data/scores-exclureSuivi.json.gz", indented)

	t.Log("/scores/liste retourne le même résultat qu'attendu avec ignoreZone=true, exclureSuivi=true et siegeUniquement=true")
	params["siegeUniquement"] = true
	_, indented, _ = post(t, "/scores/liste", params)
	processGoldenFile(t, "data/scores-siegeUniquement.json.gz", indented)
}

// TestVIAF traite la problématique du respect des traitements des droits utilisateurs
func TestVIAF(t *testing.T) {
	t.Log("absence d'etablissement vI[aA][fF]")
	if len(getSiret(t, VIAF{false, true, false, false}, 1))+
		len(getSiret(t, VIAF{false, true, false, true}, 1))+
		len(getSiret(t, VIAF{false, true, true, false}, 1))+
		len(getSiret(t, VIAF{false, true, true, true}, 1)) > 0 {
		t.Error("il existe des établissements qui ne devraient pas être là")
	}

	for _, viaf := range []string{"viaf", "viaF", "viAf", "viAF", "Viaf", "ViaF", "ViAf", "ViAF"} {
		v := VIAF{}
		v.read(viaf)

		sirets := getSiret(t, v, 1)
		if len(sirets) == 0 {
			t.Errorf("aucun siret pour tester la catégorie, test faible (%s)", viaf)
		}
		for _, siret := range sirets {
			testEtablissementVIAF(t, siret, viaf)
			testSearchVIAF(t, siret, viaf)
		}
	}
}
