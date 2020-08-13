create table regions (
  id integer primary key,
  libelle text,
  code text
);

create table departements (
  id text primary key,
  id_region integer references regions (id),
  libelle text
);

create table naf (
  id integer,
  niveau integer,
  code text,
  libelle text
);