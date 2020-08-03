create table liste (
  id int primary key,
  libelle text,
  metadata jsonb
);

create table score (
  id_liste int,
  siret text,
  siren text,
  score real,
  alerte text,
  metadata jsonb
);
