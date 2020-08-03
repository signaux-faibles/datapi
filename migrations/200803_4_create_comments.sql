create table entreprise_comments (
  id int primary key,
  id_parent int references entreprise_comments,
  siren text,
  id_keycloak text,
  date_add timestamp,
  comment text
)