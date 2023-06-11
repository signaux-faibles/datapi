create sequence if not exists campaign_id;
create table if not exists campaign (
  id integer primary key default nextval('campaign_id'),
  libelle text,
  wekan_domain_regexp text,
  date_end date,
  date_create timestamp default current_timestamp
);

create sequence if not exists campaign_etablissement_id;
create table if not exists campaign_etablissement (
  id integer primary key default nextval('campaign_etablissement_id'),
  id_campaign integer,
  siret text,
  constraint fk_campaign
    foreign key(id_campaign)
      references campaign(id)
);

create sequence if not exists campaign_etablissement_action_id;
create table if not exists campaign_etablissement_action (
  id integer primary key default nextval('campaign_etablissement_action_id'),
  id_campaign_etablissement integer,
  username text,
  date_action timestamp,
  action text,
  detail text,
  constraint fk_campaign
    foreign key(id_campaign_etablissement)
      references campaign_etablissement(id)
);

create sequence if not exists campaign_etablissement_comment_id;
create table if not exists campaign_etablissement_comment (
  id integer primary key default nextval('campaign_etablissement_comment'),
  id_campaign_etablissement integer,
  username text,
  date_comment timestamp,
  comment text,
  constraint fk_campaign
  foreign key(id_campaign_etablissement)
  references campaign_etablissement(id)
);
