create table etablissement (
  siret varchar(14) primary key,
  date_add timestamp,
  siren char(9),
  data jsonb
);

create table etablissement_archive (
  siret varchar(14),
  version int,
  date_add timestamp,
  date_archive timestamp,
  siren char(9),
  data jsonb,
  primary key (siret, version)
);