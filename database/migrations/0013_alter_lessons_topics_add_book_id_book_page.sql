ALTER TABLE lessons ADD book_id bigint DEFAULT NULL REFERENCES books ON DELETE SET NULL;
ALTER TABLE lessons ADD book_page bigint DEFAULT NULL;
ALTER TABLE lessons ADD is_teacher_excused boolean DEFAULT false;
ALTER TABLE lessons ADD lesson_attributes json DEFAULT NULL;
ALTER TABLE topics ADD book_id bigint DEFAULT NULL REFERENCES books ON DELETE SET NULL;
ALTER TABLE topics ADD book_page bigint DEFAULT NULL;