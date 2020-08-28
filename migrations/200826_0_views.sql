create view entreprise0 as
select * from entreprise where version = 0;

create view entreprise_bdf0 as
select * from entreprise_bdf where version = 0;

create view entreprise_diane0 as
select * from entreprise_diane where version = 0;

create view etablissement0 as
select * from etablissement where version = 0;

create view etablissement_apconso0 as
select * from etablissement_apconso where version = 0;

create view etablissement_apdemande0 as
select * from etablissement_apdemande where version = 0;

create view etablissement_delai0 as
select * from etablissement_delai where version = 0;

create view etablissement_periode_urssaf0 as
select * from etablissement_periode_urssaf where version = 0;

create view etablissement_procol0 as
select * from etablissement_procol where version = 0;

create view score0 as
select * from score where version = 0;

create view v_naf as
select 
  n1.id id_n1,
  n2.id id_n2,
  n3.id id_n3,
  n4.id in_n4,
  n5.id id_n5,
  n1.code code_n1,
  n2.code code_n2,
  n3.code code_n3,
  n4.code code_n4,
  n5.code code_n5,
  n1.libelle libelle_n1,
  n2.libelle libelle_n2,
  n3.libelle libelle_n3,
  n4.libelle libelle_n4,
  n5.libelle libelle_n5
from naf n5
  inner join naf n4 on n5.id_parent = n4.id
  inner join naf n3 on n4.id_parent = n3.id
  inner join naf n2 on n3.id_parent = n2.id
  inner join naf n1 on n2.id_parent = n1.id;

create view v_roles as 
select siren, array_agg(distinct departement) as roles
  from etablissement
  where version = 0
  group by siren;