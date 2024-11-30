ALTER TABLE classrooms ADD COLUMN period_uid uuid;
ALTER TABLE subjects DROP COLUMN period_uid;