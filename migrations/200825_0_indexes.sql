create index idx_etablissement_siret on etablissement using hash (siret) where version = 0;
create index idx_etablissement_siren on etablissement using hash (siren) where version = 0;
create index idx_etablissement_departement on etablissement (departement) where version = 0;
create index idx_etablissement_hash on etablissement (siret, hash) where version in (0, 1);
create index idx_etablissement_naf on etablissement (ape) where version = 0;

create index idx_entreprise_siren on entreprise using hash (siren) where version = 0;
create index idx_entreprise_hash on entreprise (siren, hash) where version in (0, 1);

create index idx_etablissement_procol_siret on etablissement_procol using hash (siret) where version = 0;
create index idx_etablissement_procol_hash on etablissement_procol (siret, hash) where version in (0, 1);

create index idx_etablissement_delai_siret on etablissement_delai using hash (siret) where version = 0;
create index idx_etablissement_delai_hash on etablissement_delai (siret, hash) where version in (0, 1);

create index idx_etablissement_periode_urssaf_siret on etablissement_periode_urssaf using hash (siret) where version = 0;
create index idx_etablissement_periode_urssaf_siren on etablissement_periode_urssaf using hash (siren) where version = 0;
create index idx_etablissement_periode_urssaf_hash on etablissement_periode_urssaf (siret, hash) where version in (0, 1);

create index idx_etablissement_apconso_siret on etablissement_apconso using hash (siret) where version = 0;
create index idx_etablissement_apconso_siren on etablissement_apconso using hash (siren) where version = 0;
create index idx_etablissement_apconso_hash on etablissement_apconso (siret, hash) where version in (0, 1);

create index idx_etablissement_apdemande_siret on etablissement_apdemande using hash (siret) where version = 0;
create index idx_etablissement_apdemande_siren on etablissement_apdemande using hash (siren) where version = 0;
create index idx_etablissement_apdemande_hash on etablissement_apdemande (siret, hash) where version in (0, 1);

create index idx_score_siret on score using hash (siret) where version = 0;
create index idx_score_siren on score using hash (siren) where version = 0;
create index idx_score_alert on score using hash (alert) where version = 0;
create index idx_score_liste on score using hash (libelle_liste) where version = 0;
create index idx_score_score on score (score) where version = 0;
create index idx_score_hash on score (siret, hash) where version in (0, 1);

create index idx_liste_libelle on liste using hash (libelle) where version = 0;
create index idx_naf_code_niveau on naf (code, niveau);
create index idx_etablissement_follow on etablissement_follow (siret, user_id, active);

create index idx_naf_code on naf using hash (code);
create index idx_naf_n1 on naf (id_n1);