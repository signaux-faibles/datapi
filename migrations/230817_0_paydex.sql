drop view if exists entreprise_paydex0;
drop table if exists entreprise_paydex;

create table entreprise_paydex
(
  id                   integer default nextval('entreprise_paydex_id'::regclass) not null
    primary key,
  siren                varchar(9),
  paydex               text,
  jours_retard         integer,
  fournisseurs         integer,
  en_cours             real,
  experiences_paiement integer,
  fpi_30               integer,
  fpi_90               integer,
  date_valeur          timestamp
);

create index idx_entreprise_paydex_siren
  on entreprise_paydex (siren);

