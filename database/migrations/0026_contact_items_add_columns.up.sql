ALTER TABLE contact_items ADD COLUMN note text default null;
ALTER TABLE contact_items ADD COLUMN birth_cert_number varchar(190) default null;
ALTER TABLE contact_items ADD COLUMN classroom_name varchar(190) default null;
ALTER TABLE contact_items ADD COLUMN parent_phone varchar(190) default null;
ALTER TABLE contact_items ADD COLUMN related_uid uuid default null;
ALTER TABLE contact_items DROP COLUMN related_uids;

ALTER TABLE contact_items ADD COLUMN created_at timestamp DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE contact_items ADD COLUMN updated_at timestamp DEFAULT CURRENT_TIMESTAMP;
