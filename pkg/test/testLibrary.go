package test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"

	"datapi/pkg/core"
	"datapi/pkg/db"
)

var apiHostAndPort = atomic.Value{}
var update = flag.Bool("overwriteGoldenFiles", false, "true pour écraser les golden files pas les réponses générées par les tests d'intégration")

// DatapiDatabaseName contient le nom de la base de données de test
const DatapiDatabaseName = "datapi_test"
const DatapiLogsDatabaseName = "datapilogs_test"

func compare(expected []byte, actual []byte) string {
	diff := difflib.UnifiedDiff{
		A:       difflib.SplitLines(string(expected)),
		B:       difflib.SplitLines(string(actual)),
		Context: 1,
	}

	text, _ := difflib.GetUnifiedDiffString(diff)
	return text
}

func loadGoldenFile(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(gzipReader)
}

func saveGoldenFile(fileName string, goldenData []byte) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	gzipWriter := gzip.NewWriter(f)
	defer gzipWriter.Close()
	_, err = gzipWriter.Write(goldenData)
	if err != nil {
		return err
	}
	err = gzipWriter.Flush()
	return err
}

// GetBodyQuietly retourne le body de la réponse sous forme de tableau de bytes et ne retourne pas d'erreur
func GetBodyQuietly(r *http.Response) []byte {
	body, err := GetBody(r)
	if err != nil {
		return []byte(fmt.Sprintf("Erreur pendant la lecture de la réponse : %s", err.Error()))
	}
	return body
}

// GetBody retourne le body de la réponse sous forme de tableau de bytes et retourne une erreur en cas de problème
func GetBody(r *http.Response) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// GetIndentedBody retourne le body de la réponse sous forme pretty
func GetIndentedBody(r *http.Response) ([]byte, error) {
	body, err := GetBody(r)
	if err != nil {
		return nil, err
	}
	var prettyBody bytes.Buffer
	err = json.Indent(&prettyBody, body, "", "  ")
	return prettyBody.Bytes(), err
}

// ProcessGoldenFile compare les golden files. (Ecrase le golden file existant si `update == true`
func ProcessGoldenFile(t *testing.T, path string, data []byte) (string, error) {
	if *update {
		err := saveGoldenFile(path, data)
		if err != nil {
			return "", err
		}
		t.Logf("mise à jour du golden file %s", path)
		return "", nil
	}
	goldenFile, err := loadGoldenFile(path)
	if err != nil {
		t.Errorf("golden file non disponible: %s", err.Error())
		return "", err
	}
	diff, err := compare(goldenFile, data), nil
	if diff != "" {
		t.Errorf("differences entre le résultat et le golden file: %s \n%s", path, diff)
	}
	return diff, err
}

// HTTPPost fonction helper pour faire du POST HTTP
func HTTPPost(t *testing.T, path string, params map[string]interface{}) *http.Response {
	jsonValue, _ := json.Marshal(params)

	request, err := http.NewRequest("POST", hostname()+path, bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Errorf("api non joignable: %s", err)
		t.Fail()
	}
	request.Header = http.Header{
		"Host":          {"go.unit"},
		"Authorization": {kcBearer},
		"Content-Type":  {"application/json"},
	}
	done, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Errorf("api non joignable: %s", err)
		t.Fail()
	}
	return done
}

// HTTPPostAndFormatBody fonction helper pour faire du POST HTTP
func HTTPPostAndFormatBody(t *testing.T, path string, params map[string]interface{}) (*http.Response, []byte, error) {
	resp := HTTPPost(t, path, params)
	indented, err := GetIndentedBody(resp)
	return resp, indented, err
}

// HTTPGet fonction helper pour faire du GET HTTP
func HTTPGet(t *testing.T, path string) *http.Response {
	request, err := http.NewRequest("GET", hostname()+path, nil)
	if err != nil {
		t.Errorf("api non joignable: %s", err)
		t.Fail()
	}
	request.Header = http.Header{
		"Host":          {"go.unit"},
		"Authorization": {kcBearer},
	}
	done, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Errorf("api non joignable: %s", err)
		t.Fail()
	}
	return done
}

// HTTPGetAndFormatBody fonction helper pour faire du GET HTTP
func HTTPGetAndFormatBody(t *testing.T, path string) (*http.Response, []byte, error) {
	resp := HTTPGet(t, path)
	indented, err := GetIndentedBody(resp)
	return resp, indented, err
}

// FollowEntreprise fonction qui suit un établissement de l'entreprise dont le siren est passé en argument
func FollowEntreprise(t *testing.T, siren string) {
	_, data, err := HTTPGetAndFormatBody(t, "/entreprise/get/"+siren)
	if err != nil {
		t.Fatalf("error when get entreprise with siren '%s' -> %s", siren, err)
	}
	entreprise := JsonToEntreprise(t, data)
	siret := entreprise.EtablissementsSummary[0].Siret
	t.Logf("will follow etablissement with siret '%s for entreprise with siren '%s'", siret, siren)
	FollowEtablissement(t, siret)
}

// FollowEtablissement fonction qui suit l'établissement dont le siret est passé en argument
func FollowEtablissement(t *testing.T, siret string) {
	params := map[string]interface{}{
		"comment":  "test",
		"category": "test",
	}
	resp := HTTPPost(t, "/follow/"+siret, params)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("le suivi a échoué: %d", resp.StatusCode)
	}
	t.Logf("nouvel établissement suivi -> '%s'", siret)
}

// JsonToEntreprise fonction qui unmarshalle un json en Entreprise
func JsonToEntreprise(t *testing.T, data []byte) core.Entreprise {
	var entreprise core.Entreprise
	err := json.Unmarshal(data, &entreprise)
	if err != nil {
		t.Fatalf("error when unmarshalling entreprise -> %s", err)
	}
	return entreprise
}

func hostname() string {
	val := apiHostAndPort.Load()
	return fmt.Sprint(val)
	//return os.Getenv("DATAPI_URL")
}

// SetHostAndPort pour renseigner l'url où sera déployée l'API
func SetHostAndPort(hostAndPort string) {
	apiHostAndPort.Store(hostAndPort)
}

// GetSiret fonction qui récupère des Sirets à partir de critères VAF (Visible / Authorized / Followed)
func GetSiret(t *testing.T, v VAF, n int) []string {
	sql := `with etablissement_data_confidentielle as (
		select distinct e.siret, et.departement
		from etablissement_periode_urssaf e
		inner join etablissement et on et.siret = e.siret
		inner join etablissement_apdemande ap on ap.siret = e.siret
		where part_patronale+part_salariale != 0
	)
	select e.siret from etablissement e
	inner join etablissement_data_confidentielle d on d.siret = e.siret
	left join v_roles r on r.siren = e.siren and roles && array['75','77','78','91','92','93','94']
	left join v_entreprise_follow f on f.siren = e.siren
	left join v_alert_entreprise a on a.siren = e.siren
	where
		e.departement != '20' and
		(r.siren is not null)=$1 and -- roles
		(a.siren is not null)=$2 and -- alert
		(f.siren is not null)=$3     -- follow
	order by e.siret
	limit $4`
	rows, err := db.Get().Query(
		context.Background(),
		sql,
		v.Visible,
		v.Alert,
		v.Followed,
		n,
	)
	defer rows.Close()
	if err != nil {
		t.Errorf("problème d'accès à la base de données: %s", err.Error())
		return nil
	}
	var sirets []string
	for rows.Next() {
		var siret string
		err := rows.Scan(&siret)
		if err != nil {
			t.Errorf("problème d'accès à la base de données: %s", err.Error())
			return nil
		}
		sirets = append(sirets, siret)
	}
	return sirets
}

// VAF encode le statut d'une entreprise selon la classification VAF
type VAF struct {
	Visible  bool
	Alert    bool
	Followed bool
}

// RazEtablissementFollowing fonction qui supprime le suivi de tous les établissements
func RazEtablissementFollowing(t *testing.T) {
	_, err := db.Get().Exec(context.Background(), "delete from etablissement_follow;")
	if err != nil {
		t.Fatalf("Erreur d'accès lors du nettoyage pre-test de la base: %s", err.Error())
	}
	t.Log("les informations de suivi des etablissements ont été supprimées")
}

// PgeTest struct contenant des données à tester pour le pge
type PgeTest struct {
	Siren           string
	HasPGE          *bool
	MustFollow      bool
	ExpectedPGE     *bool
	ExpectedPermPGE bool
}

// SelectSomeSiretsToFollow // récupérer une liste de sirets à suivre de toutes les typologies d'établissements
func SelectSomeSiretsToFollow(t *testing.T) []string {
	sirets := GetSiret(t, VAF{false, false, false}, 1)
	sirets = append(sirets, GetSiret(t, VAF{false, true, false}, 1)...)
	sirets = append(sirets, GetSiret(t, VAF{true, false, false}, 1)...)
	sirets = append(sirets, GetSiret(t, VAF{true, true, false}, 1)...)
	t.Logf("sirets to follow -> %s", sirets)
	return sirets
}

// ExclureSuivi il faut que je demande à Christophe ce que fait cette fonction en plus de la fonction `RazEtablissementFollowing`
func ExclureSuivi(t *testing.T) {
	var params = make(map[string]interface{})
	params["exclureSuivi"] = true
	params["ignoreZone"] = true
	_, indented, _ := HTTPPostAndFormatBody(t, "/scores/liste", params)
	var liste Liste
	if err := json.Unmarshal(indented, &liste); err != nil {
		t.Fatalf("ne peut pas unmarshaller la liste de scores : %s", err)
	}
	if len(liste.Scores) < 1 {
		t.Error("pas assez d'établissements, test faible")
	}

	siret := liste.Scores[0].Siret
	t.Logf("suivi de l'établissement %s", siret)
	resp := HTTPPost(t, "/follow/"+siret, map[string]interface{}{
		"comment":  "test",
		"category": "test",
	})

	if resp.StatusCode != 201 {
		t.Errorf("le suivi a échoué: %d", resp.StatusCode)
	}

	_, indented, _ = HTTPPostAndFormatBody(t, "/scores/liste", params)
	if err := json.Unmarshal(indented, &liste); err != nil {
		t.Fatalf("ne peut pas unmarshaller la liste de résultats : %s", err)
	}
	for _, l := range liste.Scores {
		if l.Followed {
			t.Fail()
		}
	}
}

// InsertPGE fonction qui suit un Etablissement de l'Entreprise dont le siren est passé en argument
func InsertPGE(t *testing.T, siren string, hasPGE *bool) {
	tx, err := db.Get().Begin(context.Background())
	if err != nil {
		t.Fatalf("something bad is happening with database when begining : %s" + err.Error())
	}

	_, err = tx.Exec(
		context.Background(),
		"insert into entreprise_pge (siren, actif) values ($1, $2)", siren, hasPGE)
	if err != nil {
		if errRollback := tx.Rollback(context.Background()); errRollback != nil {
			t.Fatalf("error rolling back. Cause : %s", err.Error())
		}
		t.Fatalf("error inserting pge [%v] for siren '%s'. Cause : %s", hasPGE, siren, err.Error())
	}
	err = tx.Commit(context.Background())
	if err != nil {
		panic("something bad is happening with database" + err.Error())
	}
}

// Liste de détection
type Liste struct {
	Scores []struct {
		Siret    string `json:"siret"`
		Followed bool   `json:"followed"`
	} `json:"scores"`
}

// Read methode de la structure VAF qui la transforme en string
func (v *VAF) Read(vaf string) {
	v.Visible = vaf[0] == 'V'
	v.Alert = vaf[1] == 'A'
	v.Followed = vaf[2] == 'F'
}

// EtablissementVAF structure encapsulant un élément de SearchVAF
type EtablissementVAF struct {
	Visible       bool `json:"visible"`
	Alert         bool `json:"alert"`
	Followed      bool `json:"followed"`
	PeriodeUrssaf struct {
		Periodes      []*time.Time `json:"periodes"`
		Cotisation    []*float64   `json:"cotisation"`
		PartPatronale []*float64   `json:"partPatronale"`
		PartSalariale []*float64   `json:"partSalariale"`
		Effectif      []*int       `json:"effectif"`
	}
	APConso   []interface{} `json:"apConso"`
	APDemande []interface{} `json:"apDemande"`
	Delai     []interface{} `json:"delai"`
}

// SearchVAF structure correspondant à la réponse JSON de l'appel à un POST HTTP sur `/etablissement/search`
type SearchVAF struct {
	Results []EtablissementVAF `json:"results"`
}

// Viperize ajoute la map passée en paramètre dans la configuration Viper
func Viperize(testConfig map[string]string) error {
	log.Println("loading test config")
	core.LoadConfig("test", "config", "migrations")

	for k, v := range testConfig {
		log.Printf("add to Viper %s : %s", k, v)
		viper.Set(k, v)
	}

	return viper.ReadInConfig()
}

// ReadXls lit le fichier excel contenu dans un zip et retourne le nombre de ligne, la première ligne ou une erreur
func ReadXls(data []byte) (int, []string, error) {
	// create and open a temporary file
	f, err := os.CreateTemp(os.TempDir(), "stats_*.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	WriteFile(f.Name(), data)
	// close and remove the temporary file at the end of the program
	defer f.Close()
	defer os.Remove(f.Name())
	xlsFile, err := excelize.OpenReader(f)
	if err != nil {
		return -1, nil, errors.Wrap(err, "erreur à la lecture du fichier xls")
	}
	// lecture des colonnes
	cols, err := xlsFile.Cols(xlsFile.GetSheetList()[0])
	if err != nil {
		return -1, nil, errors.Wrap(err, "erreur pendant la lecture des colonnes")
	}
	if !cols.Next() {
		return -1, nil, errors.Wrap(err, "erreur pas de colonne dans le fichier")
	}
	cellFromFirstCol, err := cols.Rows()
	if err != nil {
		return -1, nil, errors.Wrap(err, "erreur pendant la lecture d'une colonne")
	}
	nbRows := len(cellFromFirstCol)
	// lecture des lignes
	rows, err := xlsFile.Rows(xlsFile.GetSheetList()[0])
	if !rows.Next() {
		return -1, nil, errors.Wrap(err, "pas de lignes dans le fichier")
	}
	firstRow, err := rows.Columns()
	if err != nil {
		return -1, nil, errors.Wrap(err, "erreur pendant la lecture des colonnes")
	}
	return nbRows, firstRow, nil
}

func WriteFile(filename string, data []byte) {
	slog.Info("ecriture du fichier", slog.String("filename", filename), slog.Int("length", len(data)))
	if err := os.WriteFile(filename, data, 0600); err != nil {
		log.Fatal(err)
	}
}

func EraseAccessLogs(t *testing.T) {
	_, err := GetLogPool().Exec(context.Background(), "DELETE FROM logs")
	if err != nil {
		t.Errorf("erreur pendant la suppression des access logs : %v", err)
	}
}
