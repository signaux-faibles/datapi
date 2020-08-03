create table entreprise (
  siren char(9) primary key
  date_add timestamp
  siret_siege char(14) 
  scope text[]
  data jsonb
);

create table entreprise_archive (
  siren char(9) primary key
  date_add timestamp
  date_archive timestamp
  siret_siege char(14) 
  scope text[]
  data jsonb
);