ALTER TABLE users ADD COLUMN documents json DEFAULT NULL;
ALTER TABLE users ADD COLUMN document_files varchar(255)[] DEFAULT NULL;



ALTER TABLE message_reads ADD CONSTRAINT message_reads_unique UNIQUE (user_uid, message_uid);
