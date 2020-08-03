create table etablissement (
  siret char(14)
  date_add timestamp
  siren char(9)
  data jsonb
);

create table etablissement_archive (
  siret char(14)
  date_add timestamp
  date_archive timestamp
  siren char(9)
  data jsonb
);