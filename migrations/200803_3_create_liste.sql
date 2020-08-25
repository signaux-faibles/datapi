create sequence liste_id;
create table liste (
  id int primary key default nextval('liste_id'),
  date_add timestamp default current_timestamp,
  version int default 0,
  libelle text,
  batch text,
  algo text,
  hash string
);

create_sequence score_id;
create table score (
  id int primary key default nextval('score_id')
  date_add timestamp default current_timestamp,
  version int default 0,
  siret text,
  siren text,
  libelle_liste text,
  periode time.Time,
  score real,
  alerte text,
  hash string
);
