//go:build integration
// +build integration

package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/test"
	"github.com/stretchr/testify/assert"
	"log"
	"math/big"
	"os"
	"testing"
	"time"
)

var tuTime = time.Date(2023, 03, 10, 17, 41, 58, 651387237, time.UTC)

// TestMain : lance datapi ainsi qu'un conteneur postgres bien paramétré
// les informations de base de données doivent être identique dans :
// - le fichier de configuration de test -> test/config.toml
// - le fichier de création et d'import de données dans la base -> test/data/testData.sql.gz
// - la configuration du container
func TestMain(m *testing.M) {
	var err error
	testConfig := map[string]string{}
	testConfig["postgres"] = test.GetDatapiDbURL()
	testConfig["wekanMgoURL"] = test.GetWekanDbURL()

	apiPort := generateRandomPort()
	testConfig["bind"] = ":" + apiPort
	test.SetHostAndPort("http://localhost:" + apiPort)

	adminWhitelist := "::1, 127.0.0.1"
	testConfig["adminWhitelist"] = adminWhitelist

	err = test.Viperize(testConfig)
	if err != nil {
		log.Printf("Erreur pendant la Viperation de la config : %s", err)
	}

	test.Wait4DatapiDb(test.GetDatapiDbURL())

	err = test.Authenticate()
	if err != nil {
		log.Printf("Erreur pendant l'authentification : %s", err)
	}

	// run datapi
	core.StartDatapi()
	go initAndStartAPI()
	// time to API be ready
	time.Sleep(1 * time.Second)
	// run tests
	test.FakeTime(tuTime)
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	// on peut placer ici du code de nettoyage si nécessaire

	test.UnfakeTime()
	os.Exit(code)
}

func generateRandomPort() string {
	n, err := rand.Int(rand.Reader, big.NewInt(500))
	n.Add(n, big.NewInt(30000))
	if err != nil {
		fmt.Println("erreur pendant la génération d'un nombre aléatoire : ", err)
		return ""
	}
	return n.String()
}

func TestListes(t *testing.T) {
	t.Cleanup(func() { test.RazEtablissementFollowing(t) })

	_, body, _ := test.HTTPGetAndFormatBody(t, "/listes")
	test.ProcessGoldenFile(t, "test/data/listes.json.gz", body)
}

func TestFollow(t *testing.T) {
	t.Cleanup(func() { test.RazEtablissementFollowing(t) })

	sirets := test.SelectSomeSiretsToFollow(t)

	params := map[string]interface{}{
		"comment":  "test",
		"category": "test",
	}

	for _, siret := range sirets {
		t.Logf("suivi de l'établissement %s", siret)
		resp := test.HTTPPost(t, "/follow/"+siret, params)
		if resp.StatusCode != 201 {
			t.Errorf("le suivi a échoué: %d", resp.StatusCode)
		}
	}

	for _, siret := range sirets {
		t.Logf("suivi doublon de l'établissement %s", siret)
		resp := test.HTTPPost(t, "/follow/"+siret, params)
		if resp.StatusCode != 204 {
			t.Errorf("le doublon n'a pas été détecté correctement: %d", resp.StatusCode)
		}
	}

	t.Log("Vérification de /follow")
	_, err := db.Get().Exec(context.Background(), "update etablissement_follow set since='2020-03-01'")
	if err != nil {
		t.Errorf("erreur sql : %s", err.Error())
	}
	_, indented, _ := test.HTTPGetAndFormatBody(t, "/follow")
	test.ProcessGoldenFile(t, "test/data/follow.json.gz", indented)
}

func TestSearch(t *testing.T) {
	followEtablissementsThenCleanup(t, test.SelectSomeSiretsToFollow(t))

	// tester le retour 400 en cas de recherche trop courte
	t.Log("/etablissement/search retourne 400")
	params := map[string]interface{}{
		"search": "t",
	}
	resp := test.HTTPPost(t, "/etablissement/search", params)
	if resp.StatusCode != 400 {
		t.Errorf("mauvais status retourné: %d", resp.StatusCode)
	}

	// tester la recherche par chaine de caractères
	rows, err := db.Get().Query(context.Background(), `select distinct substring(e.siret from 1 for 3) from etablissement e
	inner join departements d on d.code = e.departement
	inner join regions r on r.id = d.id_region
	where r.libelle in ('Bourgogne-Franche-Comté', 'Auvergne-Rhône-Alpes')
	order by substring(e.siret from 1 for 3)
	limit 10
	`)
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

		params := make(map[string]interface{})
		params["search"] = siret
		params["ignoreZone"] = false
		params["ignoreRoles"] = false
		t.Logf("la recherche %s est bien de la forme attendue", siret)
		_, indented, _ := test.HTTPPostAndFormatBody(t, "/etablissement/search", params)
		goldenFilePath := fmt.Sprintf("test/data/search-%d.json.gz", i)
		test.ProcessGoldenFile(t, goldenFilePath, indented)
		i++
	}

	// tester par département
	var departements []string
	var siret string
	err = db.Get().QueryRow(
		context.Background(),
		`select array_agg(distinct departement), substring(first(siret) from 1 for 3) 
			from etablissement 
			where departement < '10' and departement != '00'`,
	).Scan(&departements, &siret)
	if err != nil {
		t.Errorf("impossible de se connecter à la base: %s", err.Error())
	}

	params = map[string]interface{}{
		"departements": departements,
		"search":       siret,
		"ignoreZone":   true,
		"ignoreRoles":  true,
	}

	t.Log("la recherche filtrée par départements est bien de la forme attendue")
	_, indented, _ := test.HTTPPostAndFormatBody(t, "/etablissement/search", params)
	goldenFilePath := "test/data/searchDepartement.json.gz"
	test.ProcessGoldenFile(t, goldenFilePath, indented)
	i++

	// tester par activité
	err = db.Get().QueryRow(
		context.Background(),
		`select substring(first(siret) from 1 for 3)
		 from etablissement e
		 inner join v_naf n on n.code_n5 = e.code_activite
		 where code_n1 in ('A', 'B', 'C')
		`,
	).Scan(&siret)
	if err != nil {
		t.Errorf("impossible de se connecter à la base: %s", err.Error())
	}

	params = map[string]interface{}{
		"activites":   []string{"A", "B", "C"},
		"search":      siret,
		"ignoreZone":  true,
		"ignoreRoles": true,
	}

	t.Log("la recherche filtrée par activites est bien de la forme attendue")
	_, indented, _ = test.HTTPPostAndFormatBody(t, "/etablissement/search", params)
	goldenFilePath = "test/data/searchActivites.json.gz"
	test.ProcessGoldenFile(t, goldenFilePath, indented)
}

// TestPGE ; cette fonction teste si le PGE est bien retourné par l'API si
// - l'entreprise a un pge actif
// - l'utilisateur connecté a bien les habilitations
//
// pour que ce test fonctionne il faut vérifier que le fichier dump
// des données de test contient bien les informations suivantes ;
// 020523337	true
// 221129065	true
// 336400422	false
func TestPGE(t *testing.T) {
	assertions := assert.New(t)

	test.RazEtablissementFollowing(t)

	tests := []test.PgeTest{
		// pge is true in entreprise_pge and has habilitation => true
		{Siren: "020523337", HasPGE: &core.True, MustFollow: core.True, ExpectedPGE: &core.True, ExpectedPermPGE: true},
		// pge is true in entreprise_pge and has no habilitation => nil
		{Siren: "221129065", HasPGE: &core.True, MustFollow: core.False, ExpectedPGE: nil, ExpectedPermPGE: false},
		// pge is false in  entreprise_pge and has no habilitation => nil
		{Siren: "336400422", HasPGE: &core.False, MustFollow: core.False, ExpectedPGE: nil, ExpectedPermPGE: false},
		// pge is false in  entreprise_pge and has habilitation => false
		{Siren: "386322594", HasPGE: &core.False, MustFollow: core.True, ExpectedPGE: &core.False, ExpectedPermPGE: true},
		// not exists in entreprise_pge and has habilitation => nil
		{Siren: "417193286", HasPGE: nil, MustFollow: core.True, ExpectedPGE: nil, ExpectedPermPGE: true},
	}
	insertPgeTests(t, tests)

	for _, pgeTest := range tests {
		t.Logf("Test PGE for entreprise '%s'", pgeTest.Siren)
		siren := pgeTest.Siren

		if pgeTest.MustFollow {
			test.FollowEntreprise(t, siren)
		}

		resp, data, _ := test.HTTPGetAndFormatBody(t, "/entreprise/get/"+siren)
		assertions.Equal(200, resp.StatusCode)
		actual := test.JsonToEntreprise(t, data)
		assertions.Equal(pgeTest.ExpectedPGE, actual.PGEActif)
		for _, etab := range actual.Etablissements {
			assertions.Equal(pgeTest.ExpectedPermPGE, etab.PermPGE)
		}
	}
}

func TestScores(t *testing.T) {
	followEtablissementsThenCleanup(t, test.SelectSomeSiretsToFollow(t))

	t.Log("/scores/liste retourne le même résultat qu'attendu")
	_, indented, _ := test.HTTPPostAndFormatBody(t, "/scores/liste", nil)
	test.ProcessGoldenFile(t, "test/data/scores.json.gz", indented)

	t.Log("/scores/liste retourne le même résultat qu'attendu avec ignoreZone=true")
	params := map[string]interface{}{
		"ignoreZone": true,
	}
	_, indented, _ = test.HTTPPostAndFormatBody(t, "/scores/liste", params)
	test.ProcessGoldenFile(t, "test/data/scores-ignoreZone.json.gz", indented)

	t.Log("/scores/liste retourne le même résultat qu'attendu avec ignoreZone=true et siegeUniquement=true")
	params["siegeUniquement"] = true
	_, indented, _ = test.HTTPPostAndFormatBody(t, "/scores/liste", params)
	test.ProcessGoldenFile(t, "test/data/scores-siegeUniquement.json.gz", indented)

	t.Log("/scores/liste traite correctement les établissements suivis")
	test.ExclureSuivi(t)
}

// TestVAF traite la problématique du respect des traitements des droits utilisateurs
// le test échantillonne des entreprises en fonction de leur statut VAF (pour visible / alert / followed) afin de tester toutes les combinaisons
// le test de base sur des golden files pour vérifier que les propriétés ne changent pas, mais va également valider la présence effective des données
// les entreprises échantillonnées disposent toutes de données confidentielles pour éviter les faux positifs.
// le sigle vaf encode le statut de l'entreprise:
// V = entreprise visible, v = entreprise non visible
// A = entreprise dont un établissement a déjà été en alerte, a = entreprise sans aucune alerte
// F = entreprise dont un établissement est suivi par l'utilisateur du test, f = enterprise non suivie
func TestVAF(t *testing.T) {
	followEtablissementsThenCleanup(t, test.SelectSomeSiretsToFollow(t))

	for _, vaf := range []string{"vaf", "vaF", "vAf", "vAF", "Vaf", "VaF", "VAf", "VAF"} {
		v := test.VAF{}
		v.Read(vaf)

		sirets := test.GetSiret(t, v, 1)
		if len(sirets) == 0 {
			t.Errorf("aucun siret pour tester la catégorie, test faible (%s)", vaf)
		}
		for _, siret := range sirets {
			testEtablissementVAF(t, siret, vaf)
			testSearchVAF(t, siret, vaf)
		}
	}
}

func TestPermissions(t *testing.T) {
	followEtablissementsThenCleanup(t, test.SelectSomeSiretsToFollow(t))

	t.Log("test de la fonction permissions")
	type input struct {
		rolesUser            []string
		rolesEntreprise      []string
		firstAlertEntreprise *string
		departement          string
		followed             bool
	}

	type expected struct {
		visible bool
		inZone  bool
		score   bool
		urssaf  bool
		dgefp   bool
		bdf     bool
		pge     bool
	}

	type useCase struct {
		description string
		test        input
		result      expected
	}

	firstAlert := "firstAlert"
	tests := []useCase{
		{"entreprise hors zone, visible, avec alerte et tous les roles",
			input{
				rolesUser:            []string{"01", "urssaf", "dgefp", "bdf", "score", "pge"},
				rolesEntreprise:      []string{"01", "02", "03"},
				firstAlertEntreprise: &firstAlert,
				departement:          "02",
				followed:             false,
			},
			expected{
				visible: true,
				inZone:  false,
				score:   true,
				urssaf:  true,
				dgefp:   true,
				bdf:     true,
				pge:     true,
			},
		},
		{"entreprise hors zone, hors visible, avec alerte et tous les roles",
			input{
				rolesUser:            []string{"05", "urssaf", "dgefp", "bdf", "score", "pge"},
				rolesEntreprise:      []string{"01", "02", "03"},
				firstAlertEntreprise: &firstAlert,
				departement:          "02",
				followed:             false,
			},
			expected{
				visible: false,
				inZone:  false,
				score:   false,
				urssaf:  false,
				dgefp:   false,
				bdf:     false,
				pge:     false,
			},
		},
		{"entreprise dans la zone, avec alerte et role score",
			input{
				rolesUser:            []string{"01", "02", "score"},
				rolesEntreprise:      []string{"02"},
				firstAlertEntreprise: &firstAlert,
				departement:          "02",
				followed:             false,
			},
			expected{
				visible: true,
				inZone:  true,
				score:   true,
				urssaf:  false,
				dgefp:   false,
				bdf:     false,
			},
		},
		{"entreprise dans la zone, avec alerte et role urssaf",
			input{
				rolesUser:            []string{"01", "02", "urssaf"},
				rolesEntreprise:      []string{"02"},
				firstAlertEntreprise: &firstAlert,
				departement:          "02",
				followed:             false,
			},
			expected{
				visible: true,
				inZone:  true,
				score:   false,
				urssaf:  true,
				dgefp:   false,
				bdf:     false,
				pge:     false,
			},
		},
		{"entreprise dans la zone, avec alerte et role pge",
			input{
				rolesUser:            []string{"01", "02", "pge"},
				rolesEntreprise:      []string{"02"},
				firstAlertEntreprise: &firstAlert,
				departement:          "02",
				followed:             false,
			},
			expected{
				visible: true,
				inZone:  true,
				score:   false,
				urssaf:  false,
				dgefp:   false,
				bdf:     false,
				pge:     true,
			},
		},
		{"entreprise dans la zone, avec alerte et role dgefp",
			input{
				rolesUser:            []string{"01", "02", "dgefp"},
				rolesEntreprise:      []string{"02"},
				firstAlertEntreprise: &firstAlert,
				departement:          "02",
				followed:             false,
			},
			expected{
				visible: true,
				inZone:  true,
				score:   false,
				urssaf:  false,
				dgefp:   true,
				bdf:     false,
				pge:     false,
			},
		},
		{"entreprise dans la zone, avec alerte et role bdf",
			input{
				rolesUser:            []string{"01", "02", "bdf"},
				rolesEntreprise:      []string{"02"},
				firstAlertEntreprise: &firstAlert,
				departement:          "02",
				followed:             false,
			},
			expected{
				visible: true,
				inZone:  true,
				score:   false,
				urssaf:  false,
				dgefp:   false,
				bdf:     true,
			},
		},
		{"entreprise dans la zone, sans alerte avec droits",
			input{
				rolesUser:            []string{"02", "urssaf", "score", "dgefp", "bdf"},
				rolesEntreprise:      []string{"01", "02", "03"},
				firstAlertEntreprise: nil,
				departement:          "02",
				followed:             false,
			},
			expected{
				visible: true,
				inZone:  true,
				score:   false,
				urssaf:  false,
				dgefp:   false,
				bdf:     false,
			},
		},
		{"entreprise hors zone, avec alerte, avec droits et suivi",
			input{
				rolesUser:            []string{"02", "urssaf", "score", "dgefp", "bdf"},
				rolesEntreprise:      []string{"05"},
				firstAlertEntreprise: &firstAlert,
				departement:          "05",
				followed:             true,
			},
			expected{
				visible: false,
				inZone:  false,
				score:   true,
				urssaf:  true,
				dgefp:   true,
				bdf:     true,
			},
		},
		{"entreprise hors zone, sans alerte, avec droits et suivi",
			input{
				rolesUser:            []string{"02", "urssaf", "score", "dgefp", "bdf"},
				rolesEntreprise:      []string{"05"},
				firstAlertEntreprise: nil,
				departement:          "05",
				followed:             true,
			},
			expected{
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
		var r expected
		t.Log(tt.description)
		err := db.Get().QueryRow(
			context.Background(),
			"select * from permissions($1, $2, $3, $4, $5)",
			tt.test.rolesUser,
			tt.test.rolesEntreprise,
			tt.test.firstAlertEntreprise,
			tt.test.departement,
			tt.test.followed,
		).Scan(&r.visible, &r.inZone, &r.score, &r.urssaf, &r.dgefp, &r.bdf, &r.pge)
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

func followEtablissementsThenCleanup(t *testing.T, sirets []string) {
	for _, siret := range sirets {
		test.FollowEtablissement(t, siret)
	}
	// a la fin du test, tout le suivi des Etablissements est supprimé
	t.Cleanup(func() { test.RazEtablissementFollowing(t) })
}

func testEtablissementVAF(t *testing.T, siret string, vaf string) {
	v := test.VAF{}
	v.Read(vaf)

	goldenFilePath := fmt.Sprintf("test/data/getEtablissement-%s-%s.json.gz", vaf, siret)
	t.Logf("l'établissement %s est bien de la forme attendue (ref %s)", siret, goldenFilePath)
	_, indented, _ := test.HTTPGetAndFormatBody(t, "/etablissement/get/"+siret)
	test.ProcessGoldenFile(t, goldenFilePath, indented)

	var e test.EtablissementVAF
	json.Unmarshal(indented, &e)
	if !(e.Visible == v.Visible && e.Followed == v.Followed) {
		t.Errorf("l'établissement %s de type %s n'a pas les propriétés requises", siret, vaf)
	}

	if !((v.Visible && v.Alert) || v.Followed) {
		if len(e.PeriodeUrssaf.Cotisation) > 0 {
			t.Errorf("Fuite de cotisations sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.PeriodeUrssaf.PartPatronale) > 0 {
			t.Errorf("Fuite de part patronale sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.PeriodeUrssaf.PartSalariale) > 0 {
			t.Errorf("Fuite de part salariale sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.Delai) > 0 {
			t.Errorf("Fuite de délai urssaf sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.APConso) > 0 {
			t.Errorf("Fuite de consommation d'activité partielle sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.APDemande) > 0 {
			t.Errorf("Fuite de demande d'activité partielle sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
	} else {
		if len(e.PeriodeUrssaf.Cotisation) == 0 {
			t.Errorf("Absence de cotisations urssaf sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.PeriodeUrssaf.PartPatronale) == 0 {
			t.Errorf("Absence de part patronale urssaf sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
		if len(e.PeriodeUrssaf.PartSalariale) == 0 {
			t.Errorf("Absence de part salariale urssaf sur l'établissement %s (ref %s)", siret, goldenFilePath)
		}
	}
}

func testSearchVAF(t *testing.T, siret string, vaf string) {
	goldenFilePath := fmt.Sprintf("test/data/getSearch-%s-%s.json.gz", vaf, siret)
	t.Logf("la recherche renvoie l'établissement %s sous la forme attendue (ref %s)", siret, goldenFilePath)
	params := map[string]interface{}{
		"search":      siret,
		"ignoreZone":  true,
		"ignoreRoles": true,
	}
	_, indented, _ := test.HTTPPostAndFormatBody(t, "/etablissement/search", params)
	diff, _ := test.ProcessGoldenFile(t, goldenFilePath, indented)
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: %s \n%s", goldenFilePath, diff)
	}
	visible := vaf[0] == 'V'
	followed := vaf[2] == 'F'
	var e test.SearchVAF
	json.Unmarshal(indented, &e)
	if len(e.Results) != 1 || !(e.Results[0].Visible == visible && e.Results[0].Followed == followed) {
		fmt.Println(vaf, visible, followed)
		t.Errorf("la recherche %s de type %s n'a pas les propriétés requises", siret, vaf)
	}
}

// insertPgeTests add pge in entreprise_pge for a siren
func insertPgeTests(t *testing.T, pgesData []test.PgeTest) {
	for _, pgeTest := range pgesData {
		if pgeTest.HasPGE == nil {
			continue
		}
		test.InsertPGE(t, pgeTest.Siren, pgeTest.HasPGE)
	}
}
