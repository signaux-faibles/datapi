# modèle de données datapi v2

## Objectif
permettre l'inclusion d'évènements sur les entreprises dans un modèle relationnel rapide.

## Modèle
### entreprise
- siren
- scope
- siret_siege
- business_data…

### entreprise_archive
- siren
- date_archive
- business_data…

### entreprise_scope
- siren
- scope

### etablissement
- siren
- siret
- business_data…

### etablissement_archive
- siret
- date_archive
- business_data…

### etablissement_scope
- siret
- scope

### etablissement_scope_archive
- siret
- scope
- date_archive

### entreprise_followers
- siren
- id_keycloak
- date
- status
- business_data…

### entreprise_followers_archive
- siren
- id_keycloak
- date
- date_archive
- status
- business_data…

### entreprise_comments
- id
- id_parent
- siren
- id_keycloak
- date
- comment

### score
- siret
- siren
- id_liste
- score
- alerte

### liste
- id_liste
- libelle

## api
### entreprises/établissements
#### GET /entreprise/{siren}
récupère les données d'une entreprise (et de l'établissement siège)
#### GET /etablissement/{siret}
récupère les données d'un établissement (et de l'entreprise parente)
#### GET /etablissements/{siren}
récupère les données de tous les établissements d'une entreprise

### suivi d'enterprise
#### GET /follow
récupère les entreprises (+établissement siège) suivies par l'utilisateur
#### POST /follow/{siren}
enregistre un suivi d'entreprise

### commentaires
#### GET /entreprise/comments/{siren}
renvoie les commentaires d'une entreprise
#### POST /entreprise/comments/{siren}
enregistre un commentaire d'une entreprise

### scores
#### GET /listes
renvoie les listes de détection
#### GET /scores
renvoie les scores (+données synthétiques entreprises) de la dernière liste
#### GET /scores/{liste}
renvoie les scores (+données synthétiques entreprises) d'une liste

### références -> à voir si on l'intègre directement dans le code en dur
#### GET /reference
renvoie naf, départements/régions,

## Données métier

### utilisées
#### score
alert
score

#### entreprise
annee_ca
ca
connu (followers)
raison_sociale
resultat_expl
variation_ca
bdf[].arrete_bilan_bdf
bdf[].delai_fournisseur
bdf[].financier_court_terme
bdf[].poids_frng
diane[].arrete_bilan_diane
diane[].ca
diane[].credit_client
diane[].resultat_expl

#### etablissement
activite
activite_partielle
departement
dernier_effectif (précalculé, issu des effectifs urssaf) 
etat_procol (faire une revue des procédures collectives)
urssaf
apdemande.effectif_autorise
apdemande.effectif_consomme
apdemande.periode.end
apdemande.periode.start
cotisation[]
debit[].part_ouvriere
debit[].part_patronale
debit[].periode
effectif[].effectif
effectif[].periode

### non utilisées
#### entreprise
variation_resultat_expl
bdf[].annee_bdf
bdf[].dette_fiscale
bdf[].frais_financier
bdf[].raison_sociale
bdf[].secteur
bdf[].siren
bdf[].taux_marge
diane[].achat_marchandises
diane[].achat_matieres_premieres
diane[].autonomie_financiere
diane[].autres_achats_charges_externes
diane[].autres_produits_charges_reprises
diane[].ca_exportation
diane[].capacite_autofinancement
diane[].capacite_remboursement
diane[].charge_exceptionnelle
diane[].charge_personnel
diane[].charges_financieres
diane[].conces_brev_et_droits_sim
diane[].consommation
diane[].couverture_ca_besoin_fdr
diane[].couverture_ca_fdr
diane[].credit_fournisseur
diane[].degre_immo_corporelle
diane[].dette_fiscale_et_sociale
diane[].dotation_amortissement
diane[].endettement
diane[].endettement_global
diane[].equilibre_financier
diane[].excedent_brut_d_exploitation
diane[].exercice_diane
diane[].exportation
diane[].financement_actif_circulant
diane[].frais_de_RetD
diane[].impot_benefice
diane[].impots_taxes
diane[].independance_financiere
diane[].interets
diane[].liquidite_generale
diane[].liquidite_reduite
diane[].marge_commerciale
diane[].nom_entreprise
diane[].nombre_etab_secondaire
diane[].nombre_filiale
diane[].nombre_mois
diane[].numero_siren
diane[].operations_commun
diane[].part_autofinancement
diane[].part_etat
diane[].part_preteur
diane[].part_salaries
diane[].participation_salaries
diane[].performance
diane[].poids_bfr_exploitation
diane[].procedure_collective
diane[].production
diane[].productivite_capital_financier
diane[].productivite_capital_investi
diane[].productivite_potentiel_production
diane[].produit_exceptionnel
diane[].produits_financiers
diane[].rendement_brut_fonds_propres
diane[].rendement_capitaux_propres
diane[].rendement_ressources_durables
diane[].rentabilite_economique
diane[].rentabilite_nette
diane[].resultat_avant_impot
diane[].rotation_stocks
diane[].statut_juridique
diane[].subventions_d_exploitation
diane[].taille_compo_groupe
diane[].taux_d_investissement_productif
diane[].taux_endettement
diane[].taux_interet_financier
diane[].taux_interet_sur_ca
diane[].taux_valeur_ajoutee
diane[].valeur_ajoutee

#### etablissement
date_procol
variation_effectif
apconso
apdemande.date_statut
apdemande.heure_consomme
apdemande.hta
apdemande.id_conso
apdemande.int
apdemande.montant
apdemande.motif_recours_se
apdemande.mta
delai.action
delai.annee_creation
delai.date_creation
delai.date_echeance
delai.denomination
delai.duree_delai
delai.indic_6m
delai.montant_echeancier
delai.numero_compte
delai.numero_contentieux
delai.stade
procedure_collective[].date_procol
procedure_collective[].etat
sirene.adresse
sirene.ape
sirene.code_postal
sirene.commune
sirene.departement
sirene.lattitude
sirene.longitude
sirene.nature_juridique
sirene.nic
sirene.nic_siege
sirene.numero_voie
sirene.region
sirene.siren
sirene.type_voie

