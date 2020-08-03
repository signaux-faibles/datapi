create table entreprise (
  siren varchar(9) primary key,
  date_add timestamp,
  siret_siege char(14),
  data jsonb
);

create table entreprise_scope (
  id integer primary key,
  siren varchar(9) references entreprise (siren),
  scope text[]
);

create table entreprise_archive (
  siren varchar(9),
  version int,
  date_add timestamp,
  date_archive timestamp,
  siret_siege char(14),
  data jsonb,
  primary key (siren, version)
);