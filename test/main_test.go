package main

import (
	"context"
	"encoding/json"
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
		diff, _ := processGoldenFile(t, goldenFilePath, indented)
		if diff != "" {
			t.Errorf("differences entre le résultat et le golden file: 'data/search-%d.json.gz' \n%s", i, diff)
		}
		i++
	}
}

func TestFollow(t *testing.T) {
	// récupérer une liste de sirets à suivre de toutes les typologies d'établissements
	sirets := getSiret(t, false, false, false, false, 4)
	sirets = append(sirets, getSiret(t, false, false, true, false, 4)...)
	sirets = append(sirets, getSiret(t, false, true, false, false, 4)...)
	sirets = append(sirets, getSiret(t, false, true, true, false, 4)...)
	sirets = append(sirets, getSiret(t, true, false, false, false, 4)...)
	sirets = append(sirets, getSiret(t, true, false, true, false, 4)...)
	sirets = append(sirets, getSiret(t, true, true, false, false, 4)...)
	sirets = append(sirets, getSiret(t, true, true, true, false, 4)...)

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
}

func testEtablissementVIAF(t *testing.T, siret string, viaf string) {
	goldenFilePath := fmt.Sprintf("data/getEtablissement-%s-%s.json.gz", viaf, siret)
	t.Logf("l'établissement %s est bien de la forme attendue (ref %s)", siret, goldenFilePath)
	_, indented, _ := get(t, "/etablissement/get/"+siret)
	diff, _ := processGoldenFile(t, goldenFilePath, indented)
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: %s \n%s", goldenFilePath, diff)
	}

	visible := viaf[0] == 'V'
	inZone := viaf[1] == 'I'
	// alert := viaf[2] == 'A'
	followed := viaf[3] == 'F'
	var e etablissementVIAF
	json.Unmarshal(indented, &e)
	if !(e.Visible == visible && e.InZone == inZone && e.Followed == followed) {
		fmt.Println(e)
		t.Errorf("l'établissement %s de type %s n'a pas les propriétés requises", siret, viaf)
	}

}

func TestVIAF(t *testing.T) {
	t.Log("absence d'etablissement vI[aA][fF]")
	if len(getSiret(t, false, true, false, false, 4))+
		len(getSiret(t, false, true, false, true, 4))+
		len(getSiret(t, false, true, true, false, 4))+
		len(getSiret(t, false, true, true, true, 4)) > 0 {
		t.Error("il existe des établissements qui ne devraient pas être là")
	}

	for _, viaf := range []string{"viaf", "viaF", "viAf", "viAF", "Viaf", "ViaF", "ViAf", "ViAF"} {
		visible := viaf[0] == 'V'
		inZone := viaf[1] == 'I'
		alert := viaf[2] == 'A'
		followed := viaf[3] == 'F'

		sirets := getSiret(t, visible, inZone, alert, followed, 4)
		for _, siret := range sirets {
			testEtablissementVIAF(t, siret, viaf)
		}
	}
}
