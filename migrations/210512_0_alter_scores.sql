alter table score add expl_selection_concerning jsonb default '[]';
alter table score add expl_selection_reassuring jsonb default '[]';
alter table score add macro_expl jsonb default '{}';
alter table score add micro_expl jsonb default '{}';
alter table score add macro_radar jsonb default '{}';
create or replace view score0 as
select * from score where version = 0;
