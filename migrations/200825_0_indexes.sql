create index idx_etablissement_siret on etablissement (siret) where version = 0;
create index idx_etablissement_siren on etablissement (siren) where version = 0;
create index idx_etablissement_departement on etablissement (departement) where version = 0;
create index idx_etablissement_naf on etablissement (ape) where version = 0;

create index idx_entreprise_siren on entreprise (siren) where version = 0;

create index idx_etablissement_procol_siret on etablissement_procol (siret) where version = 0;

create index idx_etablissement_delai_siret on etablissement_delai (siret) where version = 0;

create index idx_etablissement_periode_urssaf_siret on etablissement_periode_urssaf (siret) where version = 0;
create index idx_etablissement_periode_urssaf_siren on etablissement_periode_urssaf (siren) where version = 0;

create index idx_etablissement_apconso_siret on etablissement_apconso (siret) where version = 0;
create index idx_etablissement_apconso_siren on etablissement_apconso (siren) where version = 0;

create index idx_etablissement_apdemande_siret on etablissement_apdemande (siret) where version = 0;
create index idx_etablissement_apdemande_siren on etablissement_apdemande (siren) where version = 0;

create index idx_score_siret on score (siret) where version = 0;
create index idx_score_siren on score (siren) where version = 0;
create index idx_score_alert on score (alert) where version = 0;
create index idx_score_liste on score (libelle_liste) where version = 0;
create index idx_score_score on score (score) where version = 0;

create index idx_liste_libelle on liste (libelle) where version = 0;
create index idx_naf_code_niveau on naf (code, niveau);

create index idx_naf_code on naf (code);
create index idx_naf_n1 on naf (id_n1);