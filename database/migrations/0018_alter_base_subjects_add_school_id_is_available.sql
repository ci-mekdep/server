ALTER TABLE base_subjects ADD school_id bigint DEFAULT NULL REFERENCES schools ON DELETE SET NULL;
ALTER TABLE base_subjects ADD is_available boolean;
ALTER TABLE base_subjects ADD created_at timestamp DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE base_subjects ADD updated_at timestamp DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE base_subjects ALTER COLUMN age_category TYPE varchar(255);
ALTER TABLE books ADD created_at timestamp DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE books ADD updated_at timestamp DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE contact_items ADD created_at timestamp DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE contact_items ADD updated_at timestamp DEFAULT CURRENT_TIMESTAMP;
