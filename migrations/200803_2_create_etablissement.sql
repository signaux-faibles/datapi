create sequence etablissement_id;
create table etablissement (
  id               integer primary key default nextval('etablissement_id'),
  siret            varchar(14),
  siren            varchar(9),
  version          integer default 0,
  date_add         timestamp default current_timestamp,
  adresse          text,
  ape              text,
  code_postal      text,
  commune          text,
  departement      text,
  lattitude        real,
  longitude        real,
  nature_juridique text,
  numero_voie      text,
  region           text,
  type_voie        text,
  visite_fce       boolean,
  statut_procol    text,
  siege            boolean,
  hash             text
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
  hash             text
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
  hash                text
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
  last_periode        boolean,
  hash                text
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
  hash               text
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
  hash          text
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