
-- UNIQUES
ALTER TABLE period_grades DROP CONSTRAINT IF EXISTS "period_grades_unique";


-- pozanda bosadyar (ON DELETE SET NULL)
ALTER TABLE report_items DROP CONSTRAINT IF EXISTS report_items_updated_by_fkey;
ALTER TABLE report_items DROP CONSTRAINT IF EXISTS report_items_unique;


-- awtomat pozulyar (ON DELETE CASCADE)
ALTER TABLE report_items DROP CONSTRAINT IF EXISTS report_items_classroom_uid_fkey;
ALTER TABLE lessons DROP CONSTRAINT IF EXISTS lessons_subject_uid_fkey;




-- TODO: tertiplemeliler
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_session_id_fkey;
  
ALTER TABLE messages_reads
  DROP CONSTRAINT IF EXISTS messages_reads_session_id_fkey;

ALTER TABLE period_grades
  DROP CONSTRAINT IF EXISTS period_grades_student_id;







ALTER TABLE messages DROP CONSTRAINT messages_session_uid_fkey;

ALTER TABLE report_items DROP CONSTRAINT unique_school_classroom;


ALTER TABLE report_items DROP CONSTRAINT report_items_uid_key;
ALTER TABLE classrooms DROP CONSTRAINT classrooms_uid_key;

ALTER TABLE lessons DROP CONSTRAINT lessons_book_uid_fkey;

ALTER TABLE timetables DROP CONSTRAINT timetables_shift_uid_fkey;


ALTER TABLE classrooms DROP CONSTRAINT classrooms_period_uid_fkey; -- NO ACTION
ALTER TABLE classrooms DROP CONSTRAINT classrooms_shift_uid_fkey; -- NO ACTION
ALTER TABLE classrooms DROP CONSTRAINT classrooms_school_uid_fkey; -- NO ACTION
DELETE FROM classrooms WHERE school_uid IS NULL;


delete from subjects where school_uid is null;

ALTER TABLE subjects DROP CONSTRAINT subjects_school_uid_fkey; -- NO ACTION

ALTER TABLE subjects DROP CONSTRAINT subjects_base_subject_uid_fkey; -- NO ACTION

ALTER TABLE subjects DROP CONSTRAINT subjects_classroom_uid_fkey; -- NO ACTION

ALTER TABLE lessons DROP CONSTRAINT lessons_uid_key;

ALTER TABLE lessons DROP CONSTRAINT lessons_subject_uid_fkey; -- CASCADE

ALTER TABLE grades DROP CONSTRAINT grades_uid_key;

ALTER TABLE grades DROP CONSTRAINT grades_lesson_uid_fkey; -- CASCADE

ALTER TABLE absents DROP CONSTRAINT absents_uid_key;

ALTER TABLE absents DROP CONSTRAINT absents_lesson_uid_fkey; -- CASCADE

ALTER TABLE user_logs DROP CONSTRAINT user_logs_subject_uid_fkey; -- CASCADE




-- others

ALTER TABLE report_items DROP CONSTRAINT report_items_updated_by_fkey1;
ALTER TABLE report_items DROP CONSTRAINT report_items_updated_by_fkey2;
ALTER TABLE report_items DROP CONSTRAINT report_items_updated_by_fkey3;
ALTER TABLE report_items DROP CONSTRAINT report_items_updated_by_fkey4;
ALTER TABLE report_items DROP CONSTRAINT report_items_updated_by_fkey5;
ALTER TABLE report_items DROP CONSTRAINT report_items_updated_by_fkey6;
ALTER TABLE report_items DROP CONSTRAINT report_items_updated_by_fkey7;
ALTER TABLE report_items DROP CONSTRAINT report_items_updated_by_fkey8;
ALTER TABLE report_items DROP CONSTRAINT report_items_updated_by_fkey9;

ALTER TABLE report_items DROP CONSTRAINT report_items_classroom_uid_fkey1;
ALTER TABLE report_items DROP CONSTRAINT report_items_classroom_uid_fkey2;
ALTER TABLE report_items DROP CONSTRAINT report_items_classroom_uid_fkey3;
ALTER TABLE report_items DROP CONSTRAINT report_items_classroom_uid_fkey4;
ALTER TABLE report_items DROP CONSTRAINT report_items_classroom_uid_fkey5;
ALTER TABLE report_items DROP CONSTRAINT report_items_classroom_uid_fkey6;
ALTER TABLE report_items DROP CONSTRAINT report_items_classroom_uid_fkey7;
ALTER TABLE report_items DROP CONSTRAINT report_items_classroom_uid_fkey8;
ALTER TABLE report_items DROP CONSTRAINT report_items_classroom_uid_fkey9;



ALTER TABLE student_notes DROP CONSTRAINT student_notes_student_uid_fkey1;

ALTER TABLE classrooms DROP CONSTRAINT classrooms_period_uid_fkey1;
ALTER TABLE classrooms DROP CONSTRAINT classrooms_period_uid_fkey2;
ALTER TABLE classrooms DROP CONSTRAINT classrooms_period_uid_fkey3;
ALTER TABLE classrooms DROP CONSTRAINT classrooms_period_uid_fkey4;
ALTER TABLE classrooms DROP CONSTRAINT classrooms_period_uid_fkey5;
ALTER TABLE classrooms DROP CONSTRAINT classrooms_period_uid_fkey6;
ALTER TABLE classrooms DROP CONSTRAINT classrooms_period_uid_fkey7;
ALTER TABLE classrooms DROP CONSTRAINT classrooms_period_uid_fkey8;
ALTER TABLE classrooms DROP CONSTRAINT classrooms_period_uid_fkey9;
