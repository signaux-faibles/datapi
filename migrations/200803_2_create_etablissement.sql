create table etablissement (
  id                integer primary key,
  siret             varchar(14),
  version           integer,
  date_add          timestamp,
  date_archive      timestamp,
  siren             char(9),
  adresse           text,
  ape               text,
  code_postal       text,
  commune           text,
  departement       text,
  lattitude         text,
  longitude         text,
  nature_juridique  text,
  numero_voie       text,
  region            text,
  type_voie         text
);

create table etablissement_apconso (
  id                integer primary key,
  id_etablissement  integer references etablissement (id),  
  date_statut       date,
  heure_consomme    int,
  hta               real,
  id_conso          text,
  int               text,
  montant           real,
  motif_recours_se  text,
  mta               real
);

create table etablissement_apdemande (
  id                integer primary key,
  id_etablissement  integer references etablissement (id),
  effectif_autorise int, 
  effectif_consomme int,
  periode_start     date,
  periode_end       date
);

create table etablissement_periode_urssaf (
  id_etablissement integer references etablissement (id),
  periode          date,
  cotisation       real,
  part_patronale   real,
  part_salariale   real,
  penalite         real,
  effectif         real,
  last_periode     boolean
);

create table etablissement_delai (
  id integer primary key,
  id_etablissement integer references etablissement (id),
  action text,
  annee_creation int,
  date_creation date,
  date_echeance date,
  denomination text,
  duree_delai int,
  indic_6m text,
  montant_echeancier text,
  numero_compte text,
  numero_contentieux text,
  stade text
);

create table etablissement_procol (
  id integer primary key,
  id_etablissement integer references etablissement (id),
  date_procol date,
  etat date
);