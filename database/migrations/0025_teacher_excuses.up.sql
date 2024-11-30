CREATE TABLE teacher_excuses (
    uid uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    teacher_uid uuid NOT NULL REFERENCES users ON DELETE CASCADE,
    school_uid uuid NOT NULL REFERENCES schools ON DELETE CASCADE,
    start_date date NOT NULL,
    end_date date NOT NULL,
    reason varchar(255) NOT NULL,
    note varchar(255) DEFAULT NULL,
    document_files json DEFAULT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE teacher_excuses ADD COLUMN school_uid uuid DEFAULT NULL REFERENCES schools ON DELETE CASCADE;

-- UPDATE teacher_excuses te
--SET school_uid = us.school_uid
--FROM user_schools us
--WHERE te.teacher_uid = us.user_uid;