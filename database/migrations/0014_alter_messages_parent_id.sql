ALTER TABLE messages ADD parent_id bigint DEFAULT NULL REFERENCES messages ON DELETE SET NULL;