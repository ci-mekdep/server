CREATE TABLE school_transfers (
   uid uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
   student_uid uuid DEFAULT NULL REFERENCES users ON DELETE CASCADE,
   target_school_uid uuid DEFAULT NULL REFERENCES schools ON DELETE CASCADE,
   source_school_uid uuid DEFAULT NULL REFERENCES schools ON DELETE CASCADE,
   target_classroom_uid uuid DEFAULT NULL REFERENCES classrooms ON DELETE CASCADE,
   source_classroom_uid uuid DEFAULT NULL REFERENCES classrooms ON DELETE CASCADE,
   sender_note text DEFAULT NULL,
   sender_files json DEFAULT NULL,
   receiver_note text DEFAULT NULL,
   sent_by uuid DEFAULT NULL REFERENCES users,
   received_by uuid DEFAULT NULL REFERENCES users,
   status varchar(255) NOT NULL,
   created_at timestamp DEFAULT CURRENT_TIMESTAMP,
   updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);