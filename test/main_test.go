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

func TestFollow(t *testing.T) {
	_, err := db.Exec(context.Background(), "delete from etablissement_follow;")
	if err != nil {
		t.Errorf("Erreur d'accès lors du nettoyage pre-test de la base: %s", err.Error())
	}
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

	t.Log("Vérification de /follow")
	db.Exec(context.Background(), "update etablissement_follow set since='2020-03-01'")
	_, indented, _ := get(t, "/follow")
	processGoldenFile(t, "data/follow.json.gz", indented)
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

func TestPermissions(t *testing.T) {
	t.Log("test de la fonction permissions")
	type test struct {
		rolesUser            []string
		rolesEntreprise      []string
		firstAlertEntreprise *string
		departement          string
		followed             bool
	}

	type result struct {
		visible bool
		inZone  bool
		score   bool
		urssaf  bool
		dgefp   bool
		bdf     bool
	}

	type useCase struct {
		description string
		test        test
		result      result
	}

	firstAlert := "firstAlert"
	tests := []useCase{
		{"entreprise hors zone, visible, avec alerte et tous les roles",
			test{
				rolesUser:            []string{"01", "urssaf", "dgefp", "bdf", "score"},
				rolesEntreprise:      []string{"01", "02", "03"},
				firstAlertEntreprise: &firstAlert,
				departement:          "02",
				followed:             false,
			},
			result{
				visible: true,
				inZone:  false,
				score:   true,
				urssaf:  true,
				dgefp:   true,
				bdf:     true,
			},
		},
		{"entreprise hors zone, hors visible, avec alerte et tous les roles",
			test{
				rolesUser:            []string{"05", "urssaf", "dgefp", "bdf", "score"},
				rolesEntreprise:      []string{"01", "02", "03"},
				firstAlertEntreprise: &firstAlert,
				departement:          "02",
				followed:             false,
			},
			result{
				visible: false,
				inZone:  false,
				score:   false,
				urssaf:  false,
				dgefp:   false,
				bdf:     false,
			},
		},
		{"entreprise dans la zone, avec alerte et role score",
			test{
				rolesUser:            []string{"01", "02", "score"},
				rolesEntreprise:      []string{"02"},
				firstAlertEntreprise: &firstAlert,
				departement:          "02",
				followed:             false,
			},
			result{
				visible: true,
				inZone:  true,
				score:   true,
				urssaf:  false,
				dgefp:   false,
				bdf:     false,
			},
		},
		{"entreprise dans la zone, sans alerte avec droits",
			test{
				rolesUser:            []string{"02", "urssaf", "score", "dgefp", "bdf"},
				rolesEntreprise:      []string{"01", "02", "03"},
				firstAlertEntreprise: nil,
				departement:          "02",
				followed:             false,
			},
			result{
				visible: true,
				inZone:  true,
				score:   false,
				urssaf:  false,
				dgefp:   false,
				bdf:     false,
			},
		},
		{"entreprise hors zone, avec alerte, avec droits et suivi",
			test{
				rolesUser:            []string{"02", "urssaf", "score", "dgefp", "bdf"},
				rolesEntreprise:      []string{"05"},
				firstAlertEntreprise: &firstAlert,
				departement:          "05",
				followed:             true,
			},
			result{
				visible: false,
				inZone:  false,
				score:   true,
				urssaf:  true,
				dgefp:   true,
				bdf:     true,
			},
		},
		{"entreprise hors zone, sans alerte, avec droits et suivi",
			test{
				rolesUser:            []string{"02", "urssaf", "score", "dgefp", "bdf"},
				rolesEntreprise:      []string{"05"},
				firstAlertEntreprise: nil,
				departement:          "05",
				followed:             true,
			},
			result{
				visible: false,
				inZone:  false,
				score:   true,
				urssaf:  true,
				dgefp:   true,
				bdf:     true,
			},
		},
	}

	for _, tt := range tests {
		var r result
		t.Log(tt.description)
		err := db.QueryRow(context.Background(), "select * from permissions($1, $2, $3, $4, $5)",
			tt.test.rolesUser, tt.test.rolesEntreprise, tt.test.firstAlertEntreprise,
			tt.test.departement, tt.test.followed,
		).Scan(&r.visible, &r.inZone, &r.score, &r.urssaf, &r.dgefp, &r.bdf)
		if err != nil {
			t.Errorf("ne peut exécuter la fonction permissions: %s", err.Error())
		}
		if tt.result != r {
			t.Error("le résultat est incorrect")
			fmt.Println("attendu", tt.result)
			fmt.Println("obtenu", r)

		}
	}

}
