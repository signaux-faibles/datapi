create sequence entreprise_paydex_id;
create table entreprise_paydex (
  id                     integer primary key default nextval('entreprise_paydex_id'),
  siren                  varchar(9),
  version                integer default 0,
  date_add               timestamp default current_timestamp,
  date_valeur            date,
  nb_jours               integer,
  hash                   bytea
);

create index idx_entreprise_paydex_siren on entreprise_paydex (siren) where version = 0;
create view entreprise_paydex0 as 
select siren, date_valeur, nb_jours from entreprise_paydex where version = 0;