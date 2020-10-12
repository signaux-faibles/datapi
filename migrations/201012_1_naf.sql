drop view v_naf;

create materialized view v_naf as
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

create index idx_v_naf_n5 on v_naf (code_n5);
create index idx_v_naf_n1 on v_naf (code_n1);