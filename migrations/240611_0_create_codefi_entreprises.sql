create sequence codefi_entreprise_id;
create table codefi_entreprises (
  id int primary key default nextval('codefi_entreprise_id'),
  libelle text,
  siren text,
  code_departement text,
  date_add timestamp default current_timestamp
);

create index idx_codefi_entreprises_siren_libelle on codefi_entreprises (siren, libelle);
