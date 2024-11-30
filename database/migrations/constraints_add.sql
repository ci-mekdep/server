-- UNIQUES
ALTER TABLE period_grades ADD CONSTRAINT period_grades_unique UNIQUE (student_id, subject_id, period_key, exam_id);
ALTER TABLE school_settings ADD CONSTRAINT school_settings_unique UNIQUE (key, school_uid);
ALTER TABLE user_payments ADD CONSTRAINT user_payments_unique UNIQUE (user_uid, classroom_uid);
ALTER TABLE period_grades ADD CONSTRAINT period_grades_unique UNIQUE (subject_uid, student_uid, period_key);
ALTER TABLE absents ADD CONSTRAINT absents_unique UNIQUE (lesson_uid, student_uid);
ALTER TABLE grades ADD CONSTRAINT grades_unique UNIQUE (lesson_uid, student_uid);

-- pozanda bosadyar (ON DELETE SET NULL)
ALTER TABLE report_items ADD FOREIGN KEY (updated_by) REFERENCES users ON DELETE SET NULL;
ALTER TABLE report_items ADD CONSTRAINT report_items_unique UNIQUE (report_uid, school_uid, classroom_uid);


ALTER TABLE messages ADD CONSTRAINT messages_session_uid_fkey FOREIGN KEY (session_uid) REFERENCES sessions (uid) ON DELETE SET NULL;
ALTER TABLE lessons ADD  CONSTRAINT lessons_book_uid_fkey FOREIGN KEY (book_uid) REFERENCES books(uid) ON DELETE SET NULL;


-- elde pozmaly (nothing)

ALTER TABLE timetables ADD  CONSTRAINT timetables_shift_uid_fkey FOREIGN KEY (shift_uid) REFERENCES shifts(uid);

ALTER TABLE classrooms ADD  CONSTRAINT classrooms_school_uid_fkey FOREIGN KEY (school_uid) REFERENCES schools(uid);
ALTER TABLE classrooms ADD  CONSTRAINT classrooms_shift_uid_fkey FOREIGN KEY (shift_uid) REFERENCES shifts(uid);
ALTER TABLE classrooms ADD  CONSTRAINT classrooms_period_uid_fkey FOREIGN KEY (period_uid) REFERENCES periods(uid); 
-- ALTER TABLE classrooms ADD FOREIGN KEY (period_uid)  REFERENCES periods(uid);

ALTER TABLE subjects ADD  CONSTRAINT subjects_school_uid_fkey FOREIGN KEY (school_uid) REFERENCES schools(uid);
ALTER TABLE subjects ADD  CONSTRAINT subjects_base_subject_uid_fkey FOREIGN KEY (base_subject_uid) REFERENCES base_subjects(uid); 
ALTER TABLE subjects ADD  CONSTRAINT subjects_classroom_uid_fkey FOREIGN KEY (classroom_uid) REFERENCES classrooms(uid); 


-- awtomat pozulyar (ON DELETE CASCADE)
ALTER TABLE report_items ADD FOREIGN KEY (classroom_uid) REFERENCES classrooms ON DELETE CASCADE;
ALTER TABLE lessons ADD  CONSTRAINT lessons_subject_uid_fkey FOREIGN KEY (subject_uid) REFERENCES subjects(uid) ON DELETE CASCADE;
ALTER TABLE grades ADD  CONSTRAINT grades_lesson_uid_fkey FOREIGN KEY (lesson_uid) REFERENCES lessons(uid) ON DELETE CASCADE;
ALTER TABLE absents ADD  CONSTRAINT absents_lesson_uid_fkey FOREIGN KEY (lesson_uid) REFERENCES lessons(uid) ON DELETE CASCADE;


-- UPDATE COLUMN
ALTER TABLE lessons ALTER COLUMN title TYPE VARCHAR(512);

-- others
ALTER TABLE contact_items ALTER COLUMN message DROP NOT NULL;




-- MAKE DELETE CASCADE
-- user_classrooms.classrooms_uid
-- lessons.subject_uid
-- grades.lesson_uid
-- absents.lesson_uid

-- DROP CONSTRAINTS ON DELETE
-- classrooms
-- subjects
-- user_schools