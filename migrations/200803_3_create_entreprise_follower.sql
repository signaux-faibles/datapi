create table entreprise_followers (
  siren varchar(14) references entreprise (siren),
  id_keycloak text,
  date_add timestamp,
  status text,
  data jsonb,
  primary key (siren, id_keycloak)
);

create table entreprise_followers_archive (
  siren varchar(14) references entreprise (siren),
  version int,
  id_keycloak text,
  date_add timestamp,
  date_archive timestamp,
  status text,
  data jsonb,
  primary key (siren, id_keycloak, version)
);