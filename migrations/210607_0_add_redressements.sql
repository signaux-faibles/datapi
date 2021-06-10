alter table score add redressements text[] default '{}';
alter table score add alert_pre_redressements text;
create or replace view score0 as
select * from score where version = 0;