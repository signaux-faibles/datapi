package imports

import (
	"context"
	"datapi/pkg/utils"
	"encoding/json"
	"github.com/signaux-faibles/goSirene"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var expectedGeoSireneJSON = `
[
  {
   "Siren": "005520135",
   "Nic": "00038",
   "Siret": "00552013500038",
   "StatutDiffusionEtablissement": "O",
   "DateCreationEtablissement": "2007-04-20T00:00:00Z",
   "TrancheEffectifsEtablissement": "",
   "AnneeEffectifsEtablissement": "",
   "ActivitePrincipaleRegistreMetiersEtablissement": "",
   "DateDernierTraitementEtablissement": "2019-11-14T14:00:12Z",
   "EtablissementSiege": true,
   "NombrePeriodesEtablissement": 2,
   "ComplementAdresseEtablissement": "",
   "NumeroVoieEtablissement": "70",
   "IndiceRepetitionEtablissement": "",
   "TypeVoieEtablissement": "RUE",
   "LibelleVoieEtablissement": "DE LAUSANNE",
   "CodePostalEtablissement": "01220",
   "LibelleCommuneEtablissement": "DIVONNE-LES-BAINS",
   "LibelleCommuneEtrangerEtablissement": "",
   "DistributionSpecialeEtablissement": "",
   "CodeCommuneEtablissement": "01143",
   "CodeCedexEtablissement": "",
   "LibelleCedexEtablissement": "",
   "CodePaysEtrangerEtablissement": "",
   "LibellePaysEtrangerEtablissement": "",
   "ComplementAdresse2Etablissement": "",
   "NumeroVoie2Etablissement": "",
   "IndiceRepetition2Etablissement": "",
   "TypeVoie2Etablissement": "",
   "LibelleVoie2Etablissement": "",
   "CodePostal2Etablissement": "",
   "LibelleCommune2Etablissement": "",
   "LibelleCommuneEtranger2Etablissement": "",
   "DistributionSpeciale2Etablissement": "",
   "CodeCommune2Etablissement": "",
   "CodeCedex2Etablissement": "",
   "LibelleCedex2Etablissement": "",
   "CodePaysEtranger2Etablissement": "",
   "LibellePaysEtranger2Etablissement": "",
   "DateDebut": "2007-11-19T00:00:00Z",
   "EtatAdministratifEtablissement": "F",
   "Enseigne1Etablissement": "",
   "Enseigne2Etablissement": "",
   "Enseigne3Etablissement": "",
   "DenominationUsuelleEtablissement": "",
   "ActivitePrincipaleEtablissement": "17.1P",
   "NomenclatureActivitePrincipaleEtablissement": "NAFRev1",
   "CaractereEmployeurEtablissement": false,
   "Longitude": 6.143376,
   "Latitude": 46.357963,
   "Geo_score": 0.96,
   "Geo_type": "housenumber",
   "Geo_adresse": "70 Rue de Lausanne 01220 Divonne-les-Bains",
   "Geo_id": "01143_0490_00070",
   "Geo_ligne": "G",
   "Geo_l4": "70 RUE DE LAUSANNE",
   "Geo_l5": ""
  },
  {
   "Siren": "015550189",
   "Nic": "00094",
   "Siret": "01555018900094",
   "StatutDiffusionEtablissement": "O",
   "DateCreationEtablissement": "1993-03-17T00:00:00Z",
   "TrancheEffectifsEtablissement": "",
   "AnneeEffectifsEtablissement": "",
   "ActivitePrincipaleRegistreMetiersEtablissement": "",
   "DateDernierTraitementEtablissement": "0001-01-01T00:00:00Z",
   "EtablissementSiege": false,
   "NombrePeriodesEtablissement": 1,
   "ComplementAdresseEtablissement": "",
   "NumeroVoieEtablissement": "",
   "IndiceRepetitionEtablissement": "",
   "TypeVoieEtablissement": "LD",
   "LibelleVoieEtablissement": "CHAMPAGNE",
   "CodePostalEtablissement": "01540",
   "LibelleCommuneEtablissement": "VONNAS",
   "LibelleCommuneEtrangerEtablissement": "",
   "DistributionSpecialeEtablissement": "",
   "CodeCommuneEtablissement": "01457",
   "CodeCedexEtablissement": "",
   "LibelleCedexEtablissement": "",
   "CodePaysEtrangerEtablissement": "",
   "LibellePaysEtrangerEtablissement": "",
   "ComplementAdresse2Etablissement": "",
   "NumeroVoie2Etablissement": "",
   "IndiceRepetition2Etablissement": "",
   "TypeVoie2Etablissement": "",
   "LibelleVoie2Etablissement": "",
   "CodePostal2Etablissement": "",
   "LibelleCommune2Etablissement": "",
   "LibelleCommuneEtranger2Etablissement": "",
   "DistributionSpeciale2Etablissement": "",
   "CodeCommune2Etablissement": "",
   "CodeCedex2Etablissement": "",
   "LibelleCedex2Etablissement": "",
   "CodePaysEtranger2Etablissement": "",
   "LibellePaysEtranger2Etablissement": "",
   "DateDebut": "1993-03-17T00:00:00Z",
   "EtatAdministratifEtablissement": "F",
   "Enseigne1Etablissement": "",
   "Enseigne2Etablissement": "",
   "Enseigne3Etablissement": "",
   "DenominationUsuelleEtablissement": "",
   "ActivitePrincipaleEtablissement": "",
   "NomenclatureActivitePrincipaleEtablissement": "",
   "CaractereEmployeurEtablissement": false,
   "Longitude": 4.981674,
   "Latitude": 46.218742,
   "Geo_score": 0.95,
   "Geo_type": "street",
   "Geo_adresse": "Champagne 01540 Vonnas",
   "Geo_id": "01457_v3eait",
   "Geo_ligne": "G",
   "Geo_l4": "CHAMPAGNE",
   "Geo_l5": ""
  },
  {
   "Siren": "015550262",
   "Nic": "00057",
   "Siret": "01555026200057",
   "StatutDiffusionEtablissement": "O",
   "DateCreationEtablissement": "0001-01-01T00:00:00Z",
   "TrancheEffectifsEtablissement": "NN",
   "AnneeEffectifsEtablissement": "",
   "ActivitePrincipaleRegistreMetiersEtablissement": "",
   "DateDernierTraitementEtablissement": "2019-11-14T14:00:12Z",
   "EtablissementSiege": true,
   "NombrePeriodesEtablissement": 1,
   "ComplementAdresseEtablissement": "",
   "NumeroVoieEtablissement": "187",
   "IndiceRepetitionEtablissement": "",
   "TypeVoieEtablissement": "AV",
   "LibelleVoieEtablissement": "DE GENEVE",
   "CodePostalEtablissement": "01220",
   "LibelleCommuneEtablissement": "DIVONNE-LES-BAINS",
   "LibelleCommuneEtrangerEtablissement": "",
   "DistributionSpecialeEtablissement": "",
   "CodeCommuneEtablissement": "01143",
   "CodeCedexEtablissement": "",
   "LibelleCedexEtablissement": "",
   "CodePaysEtrangerEtablissement": "",
   "LibellePaysEtrangerEtablissement": "",
   "ComplementAdresse2Etablissement": "",
   "NumeroVoie2Etablissement": "",
   "IndiceRepetition2Etablissement": "",
   "TypeVoie2Etablissement": "",
   "LibelleVoie2Etablissement": "",
   "CodePostal2Etablissement": "",
   "LibelleCommune2Etablissement": "",
   "LibelleCommuneEtranger2Etablissement": "",
   "DistributionSpeciale2Etablissement": "",
   "CodeCommune2Etablissement": "",
   "CodeCedex2Etablissement": "",
   "LibelleCedex2Etablissement": "",
   "CodePaysEtranger2Etablissement": "",
   "LibellePaysEtranger2Etablissement": "",
   "DateDebut": "2000-12-31T00:00:00Z",
   "EtatAdministratifEtablissement": "F",
   "Enseigne1Etablissement": "",
   "Enseigne2Etablissement": "",
   "Enseigne3Etablissement": "",
   "DenominationUsuelleEtablissement": "",
   "ActivitePrincipaleEtablissement": "74.1J",
   "NomenclatureActivitePrincipaleEtablissement": "NAF1993",
   "CaractereEmployeurEtablissement": false,
   "Longitude": 6.139474,
   "Latitude": 46.354058,
   "Geo_score": 0.85,
   "Geo_type": "housenumber",
   "Geo_adresse": "187 Avenue de Genève 01220 Divonne-les-Bains",
   "Geo_id": "01143_0440_00187",
   "Geo_ligne": "G",
   "Geo_l4": "187 AVENUE DE GENEVE",
   "Geo_l5": ""
  },
  {
   "Siren": "015550882",
   "Nic": "00128",
   "Siret": "01555088200128",
   "StatutDiffusionEtablissement": "O",
   "DateCreationEtablissement": "1986-10-27T00:00:00Z",
   "TrancheEffectifsEtablissement": "",
   "AnneeEffectifsEtablissement": "",
   "ActivitePrincipaleRegistreMetiersEtablissement": "",
   "DateDernierTraitementEtablissement": "2019-11-14T14:00:12Z",
   "EtablissementSiege": false,
   "NombrePeriodesEtablissement": 4,
   "ComplementAdresseEtablissement": "",
   "NumeroVoieEtablissement": "13",
   "IndiceRepetitionEtablissement": "",
   "TypeVoieEtablissement": "RUE",
   "LibelleVoieEtablissement": "DES CRETS",
   "CodePostalEtablissement": "01000",
   "LibelleCommuneEtablissement": "BOURG-EN-BRESSE",
   "LibelleCommuneEtrangerEtablissement": "",
   "DistributionSpecialeEtablissement": "",
   "CodeCommuneEtablissement": "01053",
   "CodeCedexEtablissement": "",
   "LibelleCedexEtablissement": "",
   "CodePaysEtrangerEtablissement": "",
   "LibellePaysEtrangerEtablissement": "",
   "ComplementAdresse2Etablissement": "",
   "NumeroVoie2Etablissement": "",
   "IndiceRepetition2Etablissement": "",
   "TypeVoie2Etablissement": "",
   "LibelleVoie2Etablissement": "",
   "CodePostal2Etablissement": "",
   "LibelleCommune2Etablissement": "",
   "LibelleCommuneEtranger2Etablissement": "",
   "DistributionSpeciale2Etablissement": "",
   "CodeCommune2Etablissement": "",
   "CodeCedex2Etablissement": "",
   "LibelleCedex2Etablissement": "",
   "CodePaysEtranger2Etablissement": "",
   "LibellePaysEtranger2Etablissement": "",
   "DateDebut": "1993-12-25T00:00:00Z",
   "EtatAdministratifEtablissement": "F",
   "Enseigne1Etablissement": "",
   "Enseigne2Etablissement": "",
   "Enseigne3Etablissement": "",
   "DenominationUsuelleEtablissement": "",
   "ActivitePrincipaleEtablissement": "50.3A",
   "NomenclatureActivitePrincipaleEtablissement": "NAF1993",
   "CaractereEmployeurEtablissement": false,
   "Longitude": 5.225769,
   "Latitude": 46.216332,
   "Geo_score": 0.97,
   "Geo_type": "housenumber",
   "Geo_adresse": "13 rue des crets 01000 Bourg-en-Bresse",
   "Geo_id": "01053_1020_00013",
   "Geo_ligne": "G",
   "Geo_l4": "13 RUE DES CRETS",
   "Geo_l5": ""
  }
]`

func TestCopyFromGeoSirene(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	var expectedGeoSirene []goSirene.GeoSirene
	json.Unmarshal([]byte(expectedGeoSireneJSON), &expectedGeoSirene)
	expectedTuples := utils.Convert(expectedGeoSirene, geoSireneData)
	expectedTuplesJSON, _ := json.MarshalIndent(expectedTuples, " ", " ")
	file, _ := os.Open("test_StockEtablissement_utf8_geo.csv.gz")
	geoSireneParser := goSirene.GeoSireneParser(context.Background(), file)

	// WHEN
	copyFromGeoSirene := CopyFromGeoSirene{
		GeoSireneParser: geoSireneParser,
		Current:         new(goSirene.GeoSirene),
		Count:           new(int),
	}
	var actualTuples [][]interface{}
	for copyFromGeoSirene.Next() {
		actual, _ := copyFromGeoSirene.Values()
		actualTuples = append(actualTuples, actual)
	}
	actualTuplesJSON, _ := json.MarshalIndent(actualTuples, " ", " ")

	// THEN
	ass.Equal(string(expectedTuplesJSON), string(actualTuplesJSON))
}

// TODO: fix close file on error
func TestCopyFromGeoSirene_missing_file(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	file, _ := os.Open("missing_test_StockEtablissement_utf8_geo.csv.gz")
	geoSireneParser := goSirene.GeoSireneParser(context.Background(), file)

	// WHEN
	copyFromGeoSirene := CopyFromGeoSirene{
		GeoSireneParser: geoSireneParser,
		Current:         new(goSirene.GeoSirene),
		Count:           new(int),
	}
	ok := copyFromGeoSirene.Next()

	// THEN
	ass.False(ok)
}
