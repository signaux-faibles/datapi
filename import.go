package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"

	pgx "github.com/jackc/pgx/v4"
)

type importObject struct {
	ID    string      `json:"_id"`
	Value importValue `json:"value"`
}

type importValue struct {
	Sirets   *[]string      `json:"sirets"`
	Bdf      *[]interface{} `json:"bdf"`
	Diane    *[]diane       `json:"diane"`
	SireneUL *sireneUL      `json:"sirene_ul"`
}

func processEntreprise(fileName string, tx *pgx.Tx) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("error opening file: %s", err.Error())
		return err
	}
	defer file.Close()

	batches, wg := newBatchRunner(tx)
	unzip, err := gzip.NewReader(file)
	decoder := json.NewDecoder(unzip)

	i := 0
	for {
		var e entreprise
		err := decoder.Decode(&e)
		if err != nil {
			close(batches)
			wg.Wait()
			fmt.Printf("\033[2K\r%s terminated: %d objects inserted\n", fileName, i)
			if err == io.EOF {
				return nil
			}
			return err
		}

		if e.Value.SireneUL.Siren != "" {
			batch := e.getBatch()
			batches <- batch
			i++
			if math.Mod(float64(i), 100) == 0 {
				fmt.Printf("\033[2K\r%s: %d objects inserted", fileName, i)
			}
		}
	}

}

func processEtablissement(fileName string, tx *pgx.Tx) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("error opening file: %s", err.Error())
		return err
	}
	defer file.Close()
	if err != nil {
		return err
	}

	batches, wg := newBatchRunner(tx)
	unzip, err := gzip.NewReader(file)
	decoder := json.NewDecoder(unzip)

	i := 0
	for {
		var e etablissement
		err := decoder.Decode(&e)
		if err != nil {
			batches <- scoreToListe()
			close(batches)
			wg.Wait()
			fmt.Printf("\033[2K\r%s terminated: %d objects inserted\n", fileName, i)
			if err == io.EOF {
				return nil
			}
			return err
		}

		if len(e.ID) > 14 && e.Value.Sirene.Departement != "" {
			e.Value.Key = e.ID[len(e.ID)-14:]
			batches <- e.getBatch()
			i++
			if math.Mod(float64(i), 100) == 0 {
				fmt.Printf("\033[2K\r%s: %d objects sent to postgres", fileName, i)
			}
		}
	}
}

func prepareImport(tx *pgx.Tx) error {
	var batch pgx.Batch
	var tables []string = []string{
		"entreprise",
		"entreprise_bdf",
		"entreprise_diane",
		"etablissement",
		"etablissement_apconso",
		"etablissement_apdemande",
		"etablissement_periode_urssaf",
		"etablissement_delai",
		"etablissement_procol",
		"score",
		"liste",
	}
	for _, table := range tables {
		batch.Queue(fmt.Sprintf("update %s set version = version + 1", table))
	}
	r := (*tx).SendBatch(context.Background(), &batch)
	return r.Close()
}

func upgradeVersion(tx *pgx.Tx) error {
	var batch pgx.Batch
	tables := map[string]string{
		"entreprise": `siren, raison_sociale, statut_juridique`,
		"entreprise_bdf": `siren, arrete_bilan_bdf, annee_bdf, delai_fournisseur, financier_court_terme, 
			poids_frng, dette_fiscale, frais_financier, taux_marge`,
		"entreprise_diane": `siren, arrete_bilan_diane, chiffre_affaire, credit_client, resultat_expl, achat_marchandises,
			achat_matieres_premieres, autonomie_financiere, autres_achats_charges_externes, autres_produits_charges_reprises,		
			ca_exportation, capacite_autofinancement, capacite_remboursement, charge_exceptionnelle, charge_personnel,
			charges_financieres, conces_brev_et_droits_sim, consommation, couverture_ca_besoin_fdr, couverture_ca_fdr,
			credit_fournisseur, degre_immo_corporelle, dette_fiscale_et_sociale, dotation_amortissement, endettement,
			endettement_global, equilibre_financier, excedent_brut_d_exploitation, exercice_diane, exportation,
			financement_actif_circulant, frais_de_RetD, impot_benefice, impots_taxes, independance_financiere, interets,
			liquidite_generale, liquidite_reduite, marge_commerciale, nombre_etab_secondaire, nombre_filiale, nombre_mois,
			operations_commun, part_autofinancement, part_etat, part_preteur, part_salaries, participation_salaries,
			performance, poids_bfr_exploitation, procedure_collective, production, productivite_capital_financier, 
			productivite_capital_investi, productivite_potentiel_production, produit_exceptionnel, produits_financiers,	
			rendement_brut_fonds_propres, rendement_capitaux_propres, rendement_ressources_durables, rentabilite_economique, 
			rentabilite_nette, resultat_avant_impot, rotation_stocks, statut_juridique, subventions_d_exploitation,	taille_compo_groupe, 
			taux_d_investissement_productif, taux_endettement, taux_interet_financier, taux_interet_sur_ca, taux_valeur_ajoutee, valeur_ajoutee`,
		"etablissement": `siret, siren, adresse, ape, code_postal, commune, departement, lattitude, longitude, nature_juridique,
			numero_voie, region, type_voie, statut_procol, siege`,
		"etablissement_apconso": "siret, siren, id_conso, heure_consomme, montant, effectif, periode",
		"etablissement_apdemande": `siret, siren, id_demande, effectif_entreprise, effectif, date_statut, periode_start, periode_end, hta, mta, 
			effectif_autorise, motif_recours_se, heure_consomme, montant_consomme, effectif_consomme`,
		"etablissement_periode_urssaf": "siret, siren, periode, cotisation, part_patronale, part_salariale, montant_majorations, effectif, last_periode",
		"etablissement_delai": `siret, siren, action, annee_creation, date_creation, date_echeance, denomination, duree_delai, indic_6m, 
			montant_echeancier, numero_compte, numero_contentieux, stade`,
		"etablissement_procol": "siret, siren, date_effet, action_procol, stade_procol",
		"score":                "siret, siren, libelle_liste, periode, score, diff, alert",
		"liste":                "libelle, batch, algo",
	}

	for table, fields := range tables {
		batch.Queue(fmt.Sprintf("update %s set hash = md5(row(%s)::text) where version = 0", table, fields))
		f0 := strings.Split(fields, ",")[0]
		batch.Queue(fmt.Sprintf(`update %s t set version = version - 1 from
		(select unnest(array[t0.id, t1.id]) id from %s t0
		inner join %s t1 on t0.%s = t1.%s and t0.version = 0 and t1.version = 1 and t0.hash = t1.hash) t0
		where t0.id = t.id`, table, table, table, f0, f0))
		batch.Queue(fmt.Sprintf(`delete from %s where version = -1`, table))
	}
	b := (*tx).SendBatch(context.Background(), &batch)
	return b.Close()
}

func refreshMaterializedViews(tx *pgx.Tx) error {
	views := []string{
		"v_roles",
		"v_diane_variation_ca",
		"v_last_effectif",
		"v_last_procol",
		"v_hausse_urssaf"}
	for _, v := range views {
		fmt.Printf("\033[2K\rrefreshing %s", v)
		_, err := (*tx).Exec(context.Background(), fmt.Sprintf("refresh materialized view %s", v))
		if err != nil {
			return err
		}
	}
	fmt.Println("")
	log.Printf("success: %d views refreshed\n", len(views))
	return nil
}
