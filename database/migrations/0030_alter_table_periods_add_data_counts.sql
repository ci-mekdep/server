ALTER TABLE periods ADD COLUMN data_counts json DEFAULT NULL;
ALTER TABLE contact_items ADD COLUMN updated_by uuid DEFAULT NULL;