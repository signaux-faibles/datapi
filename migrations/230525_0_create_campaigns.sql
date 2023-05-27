create sequence campaign_id;
create table campaign (
  id integer primary key default nextval('campaign_id'),
  libelle text,
  wekan_domain_regexp text,
  date_end date,
  date_create timestamp default current_timestamp
);

create sequence campaign_etablissement_id;
create table campaign_etablissement (
  id integer primary key default nextval('campaign_etablissement_id'),
  id_campaign integer,
  siret text,
  data jsonb,
  constraint fk_campaign
    foreign key(id_campaign)
      references campaign(id)
);

create sequence campaign_etablissement_user_id;
create table campaign_etablissement_user (
  id integer primary key default nextval('campaign_etablissement_user_id'),
  id_campaign_etablissement integer,
  username text,
  date_start timestamp,
  date_end timestamp,
  action text,
  data jsonb,
  constraint fk_campaign
    foreign key(id_campaign_etablissement)
      references campaign_etablissement(id)
);

