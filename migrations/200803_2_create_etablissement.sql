create sequence etablissement_id;
create table etablissement (
  id                    integer primary key default nextval('etablissement_id'),
  version               integer default 0,
  date_add              timestamp default current_timestamp,
  siret                 varchar(14),
  siren                 varchar(9),
  siege                 boolean,
  creation              timestamp,
  complement_adresse    text,
  numero_voie           text,  
  indice_repetition     text,
  type_voie             text,
  voie                  text,
  commune               text,
  commune_etranger      text,
  distribution_speciale text,
  code_commune          text,
  code_cedex            text,
  cedex                 text,
  code_pays_etranger    text,
  pays_etranger         text,
  code_postal           text,
  departement           text,
  code_activite         text,
  nomen_activite        text,
  latitude              real,
  longitude             real,
  visite_fce            boolean,
  hash                  bytea
);

create sequence etablissement_apconso_id;
create table etablissement_apconso (
  id               integer primary key default nextval('etablissement_apconso_id'),
  siret            varchar(14),
  siren            varchar(9),
  version          integer default 0,
  date_add         timestamp default current_timestamp,
  id_conso         text,
  heure_consomme   real,
  montant          real,
  effectif         integer,
  periode          date,
  hash             bytea
);

create sequence etablissement_apdemande_id;
create table etablissement_apdemande (
  id                  integer primary key default nextval('etablissement_apdemande_id'),
  siret               varchar(14),
  siren               varchar(9),
  version             integer default 0,
  date_add            timestamp default current_timestamp,
  id_demande          text,
  effectif_entreprise integer,
  effectif            integer,
  date_statut         date,
  periode_start       date,
  periode_end         date,
  hta                 real,
  mta                 real,
  effectif_autorise   integer,
  motif_recours_se    integer,
  heure_consomme      real,
  montant_consomme    real,
  effectif_consomme   integer,
  hash                bytea
);

create sequence etablissement_periode_urssaf_id;
create table etablissement_periode_urssaf (
  id                  integer primary key default nextval('etablissement_periode_urssaf_id'),
  siret               varchar(14),
  siren               varchar(9),
  version             integer default 0,
  date_add            timestamp default current_timestamp,
  periode             date,
  cotisation          real,
  part_patronale      real,
  part_salariale      real,
  montant_majorations real,
  effectif            real,
  hash                bytea
);

create sequence etablissement_delai_id;
create table etablissement_delai (
  id                 integer primary key default nextval('etablissement_delai_id'),
  siret              varchar(14),
  siren              varchar(9),
  version            integer default 0,
  date_add           timestamp default current_timestamp,
  action             text,
  annee_creation     int,
  date_creation      date,
  date_echeance      date,
  denomination       text,
  duree_delai        int,
  indic_6m           text,
  montant_echeancier real,
  numero_compte      text,
  numero_contentieux text,
  stade              text,
  hash               bytea
);

create sequence etablissement_procol_id;
create table etablissement_procol (
  id            integer primary key default nextval('etablissement_procol_id'),
  siret         varchar(14),
  siren         varchar(9),
  version       integer default 0,
  date_add      timestamp default current_timestamp,
  date_effet    date,
  action_procol text,
  stade_procol  text,
  hash          bytea
);

create sequence etablissemment_follow_id;
create table etablissement_follow (
  id       integer primary key default nextval('etablissemment_follow_id'),
  siret    varchar(14),
  siren    varchar(9),
  username  text,
  active   boolean,
  since    timestamp,
  until    timestamp,
  comment  text  
);

create sequence etablissement_comment_id;
create table etablissement_comments (
  id              int primary key default nextval('etablissement_comment_id'),
  id_parent       int references etablissement_comments,
  siret           varchar(14),
  siren           varchar(9),
  username         text,
  date_history    timestamp[] default array[current_timestamp],
  message_history text[]
);