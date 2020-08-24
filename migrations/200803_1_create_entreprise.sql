create sequence entreprise_id;
create table entreprise (
  id integer primary key default nextval('entreprise_id'),
  siren char(9),
  version integer default 0,
  date_add timestamp default current_timestamp,
  siret_siege char(14),
  raison_sociale text,
  statut_juridique text,
  hash text
);

create sequence entreprise_bdf_id;
create table entreprise_bdf (
  id integer primary key default nextval('entreprise_bdf_id'),
  siren text,
  version integer default 0,
  date_add timestamp default current_timestamp,
  arrete_bilan_bdf date,
  annee_bdf int,
  delai_fournisseur real,
  financier_court_terme real,
  poids_frng real,
  dette_fiscale real,
  frais_financier real,
  taux_marge real,
  hash text
);

create sequence entreprise_diane_id;
create table entreprise_diane (
  id integer primary key default nextval('entreprise_diane_id'),
  siren text,
  version integer default 0,
  date_add timestamp default current_timestamp,
  arrete_bilan_diane date,
  nombre_etab_secondaire int,
  nombre_filiale int,
  nombre_mois int,
  chiffre_affaire real,
  credit_client real,
  resultat_expl real,
  achat_marchandises real,
  achat_matieres_premieres  real,
  autonomie_financiere real,
  autres_achats_charges_externes real,
  autres_produits_charges_reprises real,
  ca_exportation real,
  capacite_autofinancement real,
  capacite_remboursement real,
  charge_exceptionnelle real,
  charge_personnel real,
  charges_financieres real,
  conces_brev_et_droits_sim real,
  consommation real,
  couverture_ca_besoin_fdr real,
  couverture_ca_fdr real,
  credit_fournisseur real,
  degre_immo_corporelle real,
  dette_fiscale_et_sociale real,
  dotation_amortissement real,
  endettement real,
  endettement_global real,
  equilibre_financier real,
  excedent_brut_d_exploitation real,
  exercice_diane real,
  exportation real,
  financement_actif_circulant real,
  frais_de_RetD real,
  impot_benefice real,
  impots_taxes real,
  independance_financiere real,
  interets real,
  liquidite_generale real,
  liquidite_reduite real,
  marge_commerciale real,
  operations_commun real,
  part_autofinancement real,
  part_etat real,
  part_preteur real,
  part_salaries real,
  participation_salaries real,
  performance real,
  poids_bfr_exploitation real,
  procedure_collective boolean,
  production real,
  productivite_capital_financier real,
  productivite_capital_investi real,
  productivite_potentiel_production real,
  produit_exceptionnel real,
  produits_financiers real,
  rendement_brut_fonds_propres real,
  rendement_capitaux_propres real,
  rendement_ressources_durables real,
  rentabilite_economique real,
  rentabilite_nette real,
  resultat_avant_impot real,
  rotation_stocks real,
  statut_juridique text,
  subventions_d_exploitation real,
  taille_compo_groupe real,
  taux_d_investissement_productif real,
  taux_endettement real,
  taux_interet_financier real,
  taux_interet_sur_ca real,
  taux_valeur_ajoutee real,
  valeur_ajoutee real,
  hash text
);

create table entreprise_followers (
  id integer primary key,
  siren varchar(14),
  id_keycloak text,
  date_add timestamp,
  status text,
  data jsonb
);

create table entreprise_scope (
  id integer primary key,
  id_entreprise int references entreprise (id),
  siren varchar(9),
  scope text[]
);

create table entreprise_comments (
  id int primary key,
  id_parent int references entreprise_comments,
  siren text,
  id_keycloak text,
  date_add timestamp,
  comment text
);