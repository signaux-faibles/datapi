ALTER TABLE entreprise ADD COLUMN IF NOT EXISTS code_activite text;
ALTER TABLE entreprise ADD COLUMN IF NOT EXISTS nomen_activite text;

create or replace view entreprise0 as
select * from entreprise where version = 0;