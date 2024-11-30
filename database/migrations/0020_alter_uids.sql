-- archive script:
-- step 1: make schemas ids unique by region code and period code 
-- step 1.1: add uids to all tables
-- step 1.2: add all relation uids
-- step 1.3: ON delete set null 
-- (schools:admin_id,specialist_id,parent_id,classroms:student_id,teacher_id,shift_id,subjects:teacher_id,secondary_teacher_id,
-- subject_exams:teacher_id,head_teacher_id,shifts:updated_by,created_by,timetable:updated_by,created_by,
-- notifications:author_id,timetables:shift_id,grades:updated_by,created_by,absents:updated_by,created_by)
-- step 1.4: nullable relation uids (above columns also nullables)
-- step 1.5: update code for uids (stores, models)
-- step 1.6: update code for relation_uids (stores, models)
-- step 1.7: make primary all uids
-- step 1.8: after testing delete old ids and relation ids


-- step 2: merge all tables to archive table
-- step 3: optimize for reads archive database, forbid writing

-- PSQL commands:
-- \l - list of databases
-- \d - list of tables
-- \d - show table columns


-- convert to uuid columns
CREATE EXTENSION "uuid-ossp";

ALTER TABLE users ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE classrooms ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE periods ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE shifts ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE base_subjects ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE lessons ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE timetables ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE grades ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE absents ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE period_grades ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE user_logs ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE payment_transactions ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE message_groups ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE messages ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE messages_reads ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE notifications ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE user_notifications ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE reports ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE confirm_codes ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE report_items ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE user_schools ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE user_classrooms ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE user_parents ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE user_settings ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE student_notes ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE school_settings ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE books ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE topics ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE sms_sender ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE subject_exams ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE subjects ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE sessions ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE schools ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE assignments ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE contact_items ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE lesson_likes ADD COLUMN uid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4();

-- foreign key: user_schools user_uid
ALTER TABLE user_schools ADD COLUMN user_uid uuid;
UPDATE user_schools us SET user_uid = u.uid FROM users u WHERE u.id = us.user_id;
ALTER TABLE user_schools DROP COLUMN user_id;
ALTER TABLE user_schools ADD FOREIGN KEY (user_uid)  REFERENCES users(uid) ON DELETE CASCADE;
ALTER TABLE user_schools ALTER COLUMN user_uid SET NOT NULL;
-- foreign key: user_parents parent_uid
ALTER TABLE user_parents ADD COLUMN parent_uid uuid;
UPDATE user_parents us SET parent_uid = u.uid FROM users u WHERE u.id = us.parent_id;
ALTER TABLE user_parents DROP COLUMN parent_id;
ALTER TABLE user_parents ADD FOREIGN KEY (parent_uid)  REFERENCES users(uid) ON DELETE CASCADE;
ALTER TABLE user_parents ALTER COLUMN parent_uid SET NOT NULL;
-- foreign key: user_parents child_uid
ALTER TABLE user_parents ADD COLUMN child_uid uuid;
UPDATE user_parents us SET child_uid = u.uid FROM users u WHERE u.id = us.child_id;
ALTER TABLE user_parents DROP COLUMN child_id;
ALTER TABLE user_parents ADD FOREIGN KEY (child_uid)  REFERENCES users(uid) ON DELETE CASCADE;
ALTER TABLE user_parents ALTER COLUMN child_uid SET NOT NULL;
-- foreign key: user_settings user_uid
ALTER TABLE user_settings ADD COLUMN user_uid uuid;
UPDATE user_settings us SET user_uid = u.uid FROM users u WHERE u.id = us.user_id;
ALTER TABLE user_settings DROP COLUMN user_id;
ALTER TABLE user_settings ADD FOREIGN KEY (user_uid)  REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE user_settings ALTER COLUMN user_uid SET NOT NULL;
-- foreign key: user_notifications user_uid
ALTER TABLE user_notifications ADD COLUMN user_uid uuid;
UPDATE user_notifications us SET user_uid = u.uid FROM users u WHERE u.id = us.user_id;
ALTER TABLE user_notifications DROP COLUMN user_id;
ALTER TABLE user_notifications ADD FOREIGN KEY (user_uid)  REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE user_notifications ALTER COLUMN user_uid SET NOT NULL;
-- foreign key: user_classrooms user_uid
ALTER TABLE user_classrooms ADD COLUMN user_uid uuid;
UPDATE user_classrooms us SET user_uid = u.uid FROM users u WHERE u.id = us.user_id;
ALTER TABLE user_classrooms DROP COLUMN user_id;
ALTER TABLE user_classrooms ADD FOREIGN KEY (user_uid)  REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE user_classrooms ALTER COLUMN user_uid SET NOT NULL;
-- foreign key: student_notes student_uid
ALTER TABLE student_notes ADD COLUMN student_uid uuid;
UPDATE student_notes us SET student_uid = u.uid FROM users u WHERE u.id = us.student_id;
ALTER TABLE student_notes DROP COLUMN student_id;
ALTER TABLE student_notes ADD FOREIGN KEY (student_uid)  REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE student_notes ALTER COLUMN student_uid SET NOT NULL;
-- foreign key: timetables updated_by_uid
ALTER TABLE timetables ADD COLUMN updated_by_uid uuid;
UPDATE timetables us SET updated_by_uid = u.uid FROM users u WHERE u.id = us.updated_by;
ALTER TABLE timetables DROP COLUMN updated_by;
ALTER TABLE timetables ADD FOREIGN KEY (updated_by_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE timetables ALTER COLUMN updated_by_uid SET NOT NULL;
-- foreign key: classrooms student_uid
ALTER TABLE classrooms ADD COLUMN student_uid uuid;
UPDATE classrooms us SET student_uid = u.uid FROM users u WHERE u.id = us.student_id;
ALTER TABLE classrooms DROP COLUMN student_id;
ALTER TABLE classrooms ADD FOREIGN KEY (student_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE classrooms ALTER COLUMN student_uid SET NOT NULL;
-- foreign key: classrooms teacher_uid
ALTER TABLE classrooms ADD COLUMN teacher_uid uuid;
UPDATE classrooms us SET teacher_uid = u.uid FROM users u WHERE u.id = us.teacher_id;
ALTER TABLE classrooms DROP COLUMN teacher_id;
ALTER TABLE classrooms ADD FOREIGN KEY (teacher_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE classrooms ALTER COLUMN teacher_uid SET NOT NULL;
-- foreign key: subjects teacher_uid
ALTER TABLE subjects ADD COLUMN teacher_uid uuid;
UPDATE subjects us SET teacher_uid = u.uid FROM users u WHERE u.id = us.teacher_id;
ALTER TABLE subjects DROP COLUMN teacher_id;
ALTER TABLE subjects ADD FOREIGN KEY (teacher_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE subjects ALTER COLUMN teacher_uid SET NOT NULL;
-- foreign key: subjects second_teacher_uid
ALTER TABLE subjects ADD COLUMN second_teacher_uid uuid;
UPDATE subjects us SET second_teacher_uid = u.uid FROM users u WHERE u.id = us.second_teacher_id;
ALTER TABLE subjects DROP COLUMN second_teacher_id;
ALTER TABLE subjects ADD FOREIGN KEY (second_teacher_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE subjects ALTER COLUMN second_teacher_uid SET NOT NULL;

-- foreign key: subject_exams teacher_uid
ALTER TABLE subject_exams ADD COLUMN teacher_uid uuid;
UPDATE subject_exams us SET teacher_uid = u.uid FROM users u WHERE u.id = us.teacher_id;
ALTER TABLE subject_exams DROP COLUMN teacher_id;
ALTER TABLE subject_exams ADD FOREIGN KEY (teacher_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE subject_exams ALTER COLUMN teacher_uid SET NOT NULL;
-- foreign key: subject_exams head_teacher_uid
ALTER TABLE subject_exams ADD COLUMN head_teacher_uid uuid;
UPDATE subject_exams us SET head_teacher_uid = u.uid FROM users u WHERE u.id = us.head_teacher_id;
ALTER TABLE subject_exams DROP COLUMN head_teacher_id;
ALTER TABLE subject_exams ADD FOREIGN KEY (head_teacher_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE subject_exams ALTER COLUMN head_teacher_uid SET NOT NULL;
-- foreign key: student_notes student_uid

ALTER TABLE subject_exams ADD COLUMN member_teacher_uids uuid[];
UPDATE subject_exams us SET member_teacher_uids = array(SELECT u.uid FROM users u WHERE u.id = ANY(us.member_teacher_ids));
ALTER TABLE subject_exams DROP COLUMN member_teacher_ids;

ALTER TABLE student_notes ADD COLUMN student_uid uuid;
UPDATE student_notes us SET student_uid = u.uid FROM users u WHERE u.id = us.student_id;
ALTER TABLE student_notes DROP COLUMN student_id;
ALTER TABLE student_notes ADD FOREIGN KEY (student_uid)  REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE student_notes ALTER COLUMN student_uid SET NOT NULL;
-- foreign key: student_notes teacher_uid
ALTER TABLE student_notes ADD COLUMN teacher_uid uuid;
UPDATE student_notes us SET teacher_uid = u.uid FROM users u WHERE u.id = us.teacher_id;
ALTER TABLE student_notes DROP COLUMN teacher_id;
ALTER TABLE student_notes ADD FOREIGN KEY (teacher_uid)  REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE student_notes ALTER COLUMN teacher_uid SET NOT NULL;
-- foreign key: shifts updated_by_uid
ALTER TABLE shifts ADD COLUMN updated_by_uid uuid;
UPDATE shifts us SET updated_by_uid = u.uid FROM users u WHERE u.id = us.updated_by;
ALTER TABLE shifts DROP COLUMN updated_by;
ALTER TABLE shifts ADD FOREIGN KEY (updated_by_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE shifts ALTER COLUMN updated_by_uid SET NOT NULL;
-- foreign key: schools admin_uid
ALTER TABLE schools ADD COLUMN admin_uid uuid;
UPDATE schools us SET admin_uid = u.uid FROM users u WHERE u.id = us.admin_id;
ALTER TABLE schools DROP COLUMN admin_id;
ALTER TABLE schools ADD FOREIGN KEY (admin_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE schools ALTER COLUMN admin_uid SET NOT NULL;
-- foreign key: schools specialist_uid
ALTER TABLE schools ADD COLUMN specialist_uid uuid;
UPDATE schools us SET specialist_uid = u.uid FROM users u WHERE u.id = us.specialist_id;
ALTER TABLE schools DROP COLUMN specialist_id;
ALTER TABLE schools ADD FOREIGN KEY (specialist_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE schools ALTER COLUMN specialist_uid SET NOT NULL;
-- foreign key: period_grades student_uid
ALTER TABLE period_grades ADD COLUMN student_uid uuid;
UPDATE period_grades us SET student_uid = u.uid FROM users u WHERE u.id = us.student_id;
ALTER TABLE period_grades DROP COLUMN student_id;
ALTER TABLE period_grades ADD FOREIGN KEY (student_uid)  REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE period_grades ALTER COLUMN student_uid SET NOT NULL;

ALTER TABLE period_grades ADD COLUMN exam_uid uuid;
UPDATE period_grades us SET exam_uid = u.uid FROM subject_exams u WHERE u.id = us.exam_id;
ALTER TABLE period_grades DROP COLUMN exam_id;
ALTER TABLE period_grades ADD FOREIGN KEY (exam_uid)  REFERENCES subject_exams(uid) ON DELETE CASCADE;

ALTER TABLE period_grades ADD COLUMN subject_uid uuid;
UPDATE period_grades us SET subject_uid = u.uid FROM subjects u WHERE u.id = us.subject_id;
ALTER TABLE period_grades DROP COLUMN subject_id;
ALTER TABLE period_grades ADD FOREIGN KEY (subject_uid)  REFERENCES subjects(uid) ON DELETE CASCADE;

-- foreign key: grades student_uid
ALTER TABLE grades ADD COLUMN student_uid uuid;
UPDATE grades us SET student_uid = u.uid FROM users u WHERE u.id = us.student_id;
ALTER TABLE grades DROP COLUMN student_id;
ALTER TABLE grades ADD FOREIGN KEY (student_uid)  REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE grades ALTER COLUMN student_uid SET NOT NULL;
-- foreign key: grades created_by_uid
ALTER TABLE grades ADD COLUMN created_by_uid uuid;
UPDATE grades us SET created_by_uid = u.uid FROM users u WHERE u.id = us.created_by;
ALTER TABLE grades DROP COLUMN created_by;
ALTER TABLE grades ADD FOREIGN KEY (created_by_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE grades ALTER COLUMN created_by_uid SET NOT NULL;
-- foreign key: grades updated_by_uid
ALTER TABLE grades ADD COLUMN updated_by_uid uuid;
UPDATE grades us SET updated_by_uid = u.uid FROM users u WHERE u.id = us.updated_by;
ALTER TABLE grades DROP COLUMN updated_by;
ALTER TABLE grades ADD FOREIGN KEY (updated_by_uid)  REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE grades ALTER COLUMN updated_by_uid SET NOT NULL;
-- foreign key: absents student_uid
ALTER TABLE absents ADD COLUMN student_uid uuid;
UPDATE absents us SET student_uid = u.uid FROM users u WHERE u.id = us.student_id;
ALTER TABLE absents DROP COLUMN student_id;
ALTER TABLE absents ADD FOREIGN KEY (student_uid)  REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE absents ALTER COLUMN student_uid SET NOT NULL;
-- foreign key: absents created_by_uid
ALTER TABLE absents ADD COLUMN created_by_uid uuid;
UPDATE absents us SET created_by_uid = u.uid FROM users u WHERE u.id = us.created_by;
ALTER TABLE absents DROP COLUMN created_by;
ALTER TABLE absents ADD FOREIGN KEY (created_by_uid) REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE absents ALTER COLUMN created_by_uid SET NOT NULL;
-- foreign key: absents updated_by_uid
ALTER TABLE absents ADD COLUMN updated_by_uid uuid;
UPDATE absents us SET updated_by_uid = u.uid FROM users u WHERE u.id = us.updated_by;
ALTER TABLE absents DROP COLUMN updated_by;
ALTER TABLE absents ADD FOREIGN KEY (updated_by_uid) REFERENCES users(uid) ON DELETE SET NULL;
-- ALTER TABLE absents ALTER COLUMN updated_by_uid SET NOT NULL;
-- foreign key: lesson_likes user_uid
ALTER TABLE lesson_likes ADD COLUMN user_uid uuid;
UPDATE lesson_likes us SET user_uid = u.uid FROM users u WHERE u.id = us.user_id;
ALTER TABLE lesson_likes DROP COLUMN user_id;
ALTER TABLE lesson_likes ADD FOREIGN KEY (user_uid) REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE lesson_likes ALTER COLUMN user_uid SET NOT NULL;
-- foreign key: message_groups admin_uid
ALTER TABLE message_groups ADD COLUMN admin_uid uuid;
UPDATE message_groups us SET admin_uid = u.uid FROM users u WHERE u.id = us.admin_id;
ALTER TABLE message_groups DROP COLUMN admin_id;
ALTER TABLE message_groups ADD FOREIGN KEY (admin_uid) REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE message_groups ALTER COLUMN admin_uid SET NOT NULL;
-- foreign key: messages_reads user_uid
ALTER TABLE messages_reads ADD COLUMN user_uid uuid;
UPDATE messages_reads us SET user_uid = u.uid FROM users u WHERE u.id = us.user_id;
ALTER TABLE messages_reads DROP COLUMN user_id;
ALTER TABLE messages_reads ADD FOREIGN KEY (user_uid) REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE messages_reads ALTER COLUMN user_uid SET NOT NULL;
-- foreign key: messages user_uid
ALTER TABLE messages ADD COLUMN user_uid uuid;
UPDATE messages us SET user_uid = u.uid FROM users u WHERE u.id = us.user_id;
ALTER TABLE messages DROP COLUMN user_id;
ALTER TABLE messages ADD FOREIGN KEY (user_uid) REFERENCES users(uid) ON DELETE CASCADE;

ALTER TABLE messages ADD COLUMN session_uid uuid;
UPDATE messages us SET session_uid = s.uid FROM sessions s WHERE s.id = us.session_id;
ALTER TABLE messages DROP COLUMN session_id;
ALTER TABLE messages ADD FOREIGN KEY (session_uid) REFERENCES sessions(uid) ON DELETE CASCADE;
-- ALTER TABLE messages ALTER COLUMN user_uid SET NOT NULL;
-- foreign key: notifications author_uid
ALTER TABLE notifications ADD COLUMN author_uid uuid;
UPDATE notifications us SET author_uid = u.uid FROM users u WHERE u.id = us.author_id;
ALTER TABLE notifications DROP COLUMN author_id;
ALTER TABLE notifications ADD FOREIGN KEY (author_uid)  REFERENCES users(uid) ON DELETE SET NULL;
ALTER TABLE notifications ALTER COLUMN author_uid SET NOT NULL;

ALTER TABLE notifications ADD COLUMN user_uids uuid[];
UPDATE notifications n SET user_uids = ARRAY(SELECT u.uid FROM users u WHERE u.uid = ANY(n.user_uids));
ALTER TABLE notifications DROP COLUMN user_ids;

ALTER TABLE notifications ADD COLUMN school_uids uuid[];
UPDATE notifications us SET school_uids = ARRAY(SELECT s.uid FROM schools s WHERE s.id = ANY(us.school_ids));
ALTER TABLE notifications DROP COLUMN school_ids;

-- foreign key: payment_transactions payer_uid
ALTER TABLE payment_transactions ADD COLUMN payer_uid uuid;
UPDATE payment_transactions us SET payer_uid = u.uid FROM users u WHERE u.id = us.payer_id;
ALTER TABLE payment_transactions DROP COLUMN payer_id;
ALTER TABLE payment_transactions ADD FOREIGN KEY (payer_uid)  REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE payment_transactions ALTER COLUMN payer_uid SET NOT NULL;

ALTER TABLE payment_transactions ADD COLUMN user_uids uuid[];
UPDATE payment_transactions us SET user_uids = ARRAY(SELECT u.uid FROM users u WHERE u.id = ANY(us.user_ids));
ALTER TABLE payment_transactions DROP COLUMN user_ids;

-- foreign key: base_subjects school_uid
ALTER TABLE base_subjects ADD COLUMN school_uid uuid;
UPDATE base_subjects tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE base_subjects DROP COLUMN school_id;
ALTER TABLE base_subjects ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE base_subjects ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: classrooms school_uid
ALTER TABLE classrooms ADD COLUMN school_uid uuid;
UPDATE classrooms tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE classrooms DROP COLUMN school_id;
ALTER TABLE classrooms ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid);
-- ALTER TABLE classrooms ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: contact_items school_uid
ALTER TABLE contact_items ADD COLUMN school_uid uuid;
UPDATE contact_items tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE contact_items DROP COLUMN school_id;
ALTER TABLE contact_items ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE contact_items ALTER COLUMN school_uid SET NOT NULL;

ALTER TABLE contact_items ADD COLUMN user_uid uuid;
UPDATE contact_items tt SET user_uid = s.uid FROM users s WHERE s.id = tt.user_id;
ALTER TABLE contact_items DROP COLUMN user_id;
ALTER TABLE contact_items ADD FOREIGN KEY (user_uid)  REFERENCES users(uid) ON DELETE CASCADE;
-- ALTER TABLE contact_items ALTER COLUMN user_uid SET NOT NULL;

ALTER TABLE contact_items ADD COLUMN related_uids uuid[];
ALTER TABLE contact_items DROP COLUMN related_ids;

-- foreign key: message_groups school_uid
ALTER TABLE message_groups ADD COLUMN school_uid uuid;
UPDATE message_groups tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE message_groups DROP COLUMN school_id;
ALTER TABLE message_groups ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE message_groups ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: lessons school_uid
ALTER TABLE lessons ADD COLUMN school_uid uuid;
UPDATE lessons tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE lessons DROP COLUMN school_id;
ALTER TABLE lessons ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE lessons ALTER COLUMN school_uid SET NOT NULL;

ALTER TABLE lessons ADD COLUMN book_uid uuid;
UPDATE lessons tt SET book_uid = b.uid FROM books b WHERE b.id = tt.book_id;
ALTER TABLE lessons DROP COLUMN book_id;
ALTER TABLE lessons ADD FOREIGN KEY (book_uid)  REFERENCES books(uid) ON DELETE CASCADE;

-- foreign key: payment_transactions school_uid
ALTER TABLE payment_transactions ADD COLUMN school_uid uuid;
UPDATE payment_transactions tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE payment_transactions DROP COLUMN school_id;
ALTER TABLE payment_transactions ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE payment_transactions ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: periods school_uid
ALTER TABLE periods ADD COLUMN school_uid uuid;
UPDATE periods tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE periods DROP COLUMN school_id;
ALTER TABLE periods ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE periods ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: report_items school_uid
ALTER TABLE report_items ADD COLUMN school_uid uuid;
UPDATE report_items tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE report_items DROP COLUMN school_id;
ALTER TABLE report_items ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE report_items ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: school_settings school_uid
ALTER TABLE school_settings ADD COLUMN school_uid uuid;
UPDATE school_settings tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE school_settings DROP COLUMN school_id;
ALTER TABLE school_settings ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE school_settings ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: schools parent_uid
ALTER TABLE schools ADD COLUMN parent_uid uuid;
UPDATE schools tt SET parent_uid = s.uid FROM schools s WHERE s.id = tt.parent_id;
ALTER TABLE schools DROP COLUMN parent_id;
ALTER TABLE schools ADD FOREIGN KEY (parent_uid)  REFERENCES schools(uid) ON DELETE SET NULL;
-- ALTER TABLE schools ALTER COLUMN parent_uid SET NOT NULL;
-- foreign key: student_notes school_uid
ALTER TABLE student_notes ADD COLUMN school_uid uuid;
UPDATE student_notes tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE student_notes DROP COLUMN school_id;
ALTER TABLE student_notes ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE student_notes ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: shifts school_uid
ALTER TABLE shifts ADD COLUMN school_uid uuid;
UPDATE shifts tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE shifts DROP COLUMN school_id;
ALTER TABLE shifts ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE shifts ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: subjects school_uid
ALTER TABLE subjects ADD COLUMN school_uid uuid;
UPDATE subjects tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE subjects DROP COLUMN school_id;
ALTER TABLE subjects ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid);
-- ALTER TABLE subjects ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: timetables school_uid
ALTER TABLE timetables ADD COLUMN school_uid uuid;
UPDATE timetables tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE timetables DROP COLUMN school_id;
ALTER TABLE timetables ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid);
-- ALTER TABLE timetables ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: subject_exams school_uid
ALTER TABLE subject_exams ADD COLUMN school_uid uuid;
UPDATE subject_exams tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE subject_exams DROP COLUMN school_id;
ALTER TABLE subject_exams ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE subject_exams ALTER COLUMN school_uid SET NOT NULL;
-- foreign key: user_schools school_uid
ALTER TABLE user_schools ADD COLUMN school_uid uuid;
UPDATE user_schools tt SET school_uid = s.uid FROM schools s WHERE s.id = tt.school_id;
ALTER TABLE user_schools DROP COLUMN school_id;
ALTER TABLE user_schools ADD FOREIGN KEY (school_uid)  REFERENCES schools(uid) ON DELETE CASCADE;
-- ALTER TABLE user_schools ALTER COLUMN school_uid SET NOT NULL;


-- foreign key: subjects classroom_uid
ALTER TABLE subjects ADD COLUMN classroom_uid uuid;
UPDATE subjects tt SET classroom_uid = c.uid FROM classrooms c WHERE c.id = tt.classroom_id;
ALTER TABLE subjects DROP COLUMN classroom_id;
ALTER TABLE subjects ADD FOREIGN KEY (classroom_uid)  REFERENCES classrooms(uid);
-- ALTER TABLE subjects ALTER COLUMN classroom_uid SET NOT NULL;
-- foreign key: timetables classroom_uid
ALTER TABLE timetables ADD COLUMN classroom_uid uuid;
UPDATE timetables tt SET classroom_uid = c.uid FROM classrooms c WHERE c.id = tt.classroom_id;
ALTER TABLE timetables DROP COLUMN classroom_id;
ALTER TABLE timetables ADD FOREIGN KEY (classroom_uid)  REFERENCES classrooms(uid) ON DELETE CASCADE;
-- ALTER TABLE timetables ALTER COLUMN classroom_uid SET NOT NULL;
-- foreign key: user_classrooms classroom_uid
ALTER TABLE user_classrooms ADD COLUMN classroom_uid uuid;
UPDATE user_classrooms tt SET classroom_uid = c.uid FROM classrooms c WHERE c.id = tt.classroom_id;
ALTER TABLE user_classrooms DROP COLUMN classroom_id;
ALTER TABLE user_classrooms ADD FOREIGN KEY (classroom_uid)  REFERENCES classrooms(uid) ON DELETE CASCADE;
-- ALTER TABLE user_classrooms ALTER COLUMN classroom_uid SET NOT NULL;
-- foreign key: subject_exams classroom_uid
ALTER TABLE subject_exams ADD COLUMN classroom_uid uuid;
UPDATE subject_exams tt SET classroom_uid = c.uid FROM classrooms c WHERE c.id = tt.classroom_id;
ALTER TABLE subject_exams DROP COLUMN classroom_id;
ALTER TABLE subject_exams ADD FOREIGN KEY (classroom_uid)  REFERENCES classrooms(uid) ON DELETE CASCADE;
-- ALTER TABLE subject_exams ALTER COLUMN classroom_uid SET NOT NULL;
-- foreign key: message_groups classroom_uid
ALTER TABLE message_groups ADD COLUMN classroom_uid uuid;
UPDATE message_groups tt SET classroom_uid = c.uid FROM classrooms c WHERE c.id = tt.classroom_id;
ALTER TABLE message_groups DROP COLUMN classroom_id;
ALTER TABLE message_groups ADD FOREIGN KEY (classroom_uid)  REFERENCES classrooms(uid) ON DELETE CASCADE;
-- ALTER TABLE message_groups ALTER COLUMN classroom_uid SET NOT NULL;
-- foreign key: classrooms parent_uid
ALTER TABLE classrooms ADD COLUMN parent_uid uuid;
UPDATE classrooms tt SET parent_uid = c.uid FROM classrooms c WHERE c.id = tt.parent_id;
ALTER TABLE classrooms DROP COLUMN parent_id;
ALTER TABLE classrooms ADD FOREIGN KEY (parent_uid)  REFERENCES classrooms(uid) ON DELETE CASCADE;
-- ALTER TABLE classrooms ALTER COLUMN parent_uid SET NOT NULL;

-- foreign key: lessons subject_uid
ALTER TABLE lessons ADD COLUMN subject_uid uuid;
UPDATE lessons tt SET subject_uid = c.uid FROM subjects c WHERE c.id = tt.subject_id;
ALTER TABLE lessons DROP COLUMN subject_id;
ALTER TABLE lessons ADD FOREIGN KEY (subject_uid)  REFERENCES subjects(uid) ON DELETE CASCADE;
-- ALTER TABLE lessons ALTER COLUMN subject_uid SET NOT NULL;
-- foreign key: subject_exams subject_uid
ALTER TABLE subject_exams ADD COLUMN subject_uid uuid;
UPDATE subject_exams tt SET subject_uid = c.uid FROM subjects c WHERE c.id = tt.subject_id;
ALTER TABLE subject_exams DROP COLUMN subject_id;
ALTER TABLE subject_exams ADD FOREIGN KEY (subject_uid)  REFERENCES subjects(uid) ON DELETE CASCADE;
-- ALTER TABLE subject_exams ALTER COLUMN subject_uid SET NOT NULL;
-- foreign key: student_notes subject_uid
ALTER TABLE student_notes ADD COLUMN subject_uid uuid;
UPDATE student_notes tt SET subject_uid = c.uid FROM subjects c WHERE c.id = tt.subject_id;
ALTER TABLE student_notes DROP COLUMN subject_id;
ALTER TABLE student_notes ADD FOREIGN KEY (subject_uid)  REFERENCES subjects(uid) ON DELETE CASCADE;
-- ALTER TABLE student_notes ALTER COLUMN subject_uid SET NOT NULL;
-- foreign key: subjects parent_uid
ALTER TABLE subjects ADD COLUMN parent_uid uuid;
UPDATE subjects tt SET parent_uid = c.uid FROM subjects c WHERE c.id = tt.parent_id;
ALTER TABLE subjects DROP COLUMN parent_id;
ALTER TABLE subjects ADD FOREIGN KEY (parent_uid)  REFERENCES subjects(uid) ON DELETE SET NULL;
-- ALTER TABLE subjects ALTER COLUMN parent_uid SET NOT NULL;



-- foreign key: lessons period_uid
ALTER TABLE lessons ADD COLUMN period_uid uuid;
UPDATE lessons tt SET period_uid = c.uid FROM subjects c WHERE c.id = tt.period_id;
ALTER TABLE lessons DROP COLUMN period_id;
ALTER TABLE lessons ADD FOREIGN KEY (period_uid)  REFERENCES periods(uid) ON DELETE SET NULL;
-- ALTER TABLE lessons ALTER COLUMN period_uid SET NOT NULL;
-- foreign key: period_grades period_uid
ALTER TABLE period_grades ADD COLUMN period_uid uuid;
ALTER TABLE period_grades DROP COLUMN period_id;
UPDATE period_grades tt SET period_uid = c.uid FROM subjects c WHERE c.id = tt.period_id;
ALTER TABLE period_grades ADD FOREIGN KEY (period_uid)  REFERENCES periods(uid) ON DELETE SET NULL;
-- ALTER TABLE period_grades ALTER COLUMN period_uid SET NOT NULL;
-- foreign key: subjects period_uid
ALTER TABLE subjects ADD COLUMN period_uid uuid;
UPDATE subjects tt SET period_uid = c.uid FROM subjects c WHERE c.id = tt.period_id;
ALTER TABLE subjects DROP COLUMN period_id;
ALTER TABLE subjects ADD FOREIGN KEY (period_uid)  REFERENCES periods(uid) ON DELETE SET NULL;
-- ALTER TABLE subjects ALTER COLUMN period_uid SET NOT NULL;
-- foreign key: timetables period_uid
ALTER TABLE timetables ADD COLUMN period_uid uuid;
UPDATE timetables tt SET period_uid = c.uid FROM subjects c WHERE c.id = tt.period_id;
ALTER TABLE timetables DROP COLUMN period_id;
ALTER TABLE timetables ADD FOREIGN KEY (period_uid)  REFERENCES periods(uid) ON DELETE SET NULL;
-- ALTER TABLE timetables ALTER COLUMN period_uid SET NOT NULL;
-- foreign key: report_items period_uid
ALTER TABLE report_items ADD COLUMN period_uid uuid;
UPDATE report_items tt SET period_uid = c.uid FROM subjects c WHERE c.id = tt.period_id;
ALTER TABLE report_items DROP COLUMN period_id;
ALTER TABLE report_items ADD FOREIGN KEY (period_uid)  REFERENCES periods(uid) ON DELETE SET NULL;
-- ALTER TABLE report_items ALTER COLUMN period_uid SET NOT NULL;



-- foreign key: classrooms shift_uid
ALTER TABLE classrooms ADD COLUMN shift_uid uuid;
UPDATE classrooms tt SET shift_uid = c.uid FROM subjects c WHERE c.id = tt.shift_id;
ALTER TABLE classrooms DROP COLUMN shift_id;
ALTER TABLE classrooms ADD FOREIGN KEY (shift_uid)  REFERENCES shifts(uid) ON DELETE SET NULL;
-- ALTER TABLE classrooms ALTER COLUMN shift_uid SET NOT NULL;
-- foreign key: timetables shift_uid
ALTER TABLE timetables ADD COLUMN shift_uid uuid;
UPDATE timetables tt SET shift_uid = c.uid FROM subjects c WHERE c.id = tt.shift_id;
ALTER TABLE timetables DROP COLUMN shift_id;
ALTER TABLE timetables ADD FOREIGN KEY (shift_uid)  REFERENCES shifts(uid) ON DELETE SET NULL;
-- ALTER TABLE timetables ALTER COLUMN shift_uid SET NOT NULL;



-- foreign key: subjects base_subject_uid
ALTER TABLE subjects ADD COLUMN base_subject_uid uuid;
UPDATE subjects tt SET base_subject_uid = c.uid FROM subjects c WHERE c.id = tt.base_subject_id;
ALTER TABLE subjects DROP COLUMN base_subject_id;
ALTER TABLE subjects ADD FOREIGN KEY (base_subject_uid)  REFERENCES base_subjects(uid) ON DELETE SET NULL;
-- ALTER TABLE subjects ALTER COLUMN base_subject_uid SET NOT NULL;



-- foreign key: grades lesson_uid
ALTER TABLE grades ADD COLUMN lesson_uid uuid;
UPDATE grades tt SET lesson_uid = c.uid FROM subjects c WHERE c.id = tt.lesson_id;
ALTER TABLE grades DROP COLUMN lesson_id;
ALTER TABLE grades ADD FOREIGN KEY (lesson_uid)  REFERENCES lessons(uid) ON DELETE SET NULL;
-- ALTER TABLE grades ALTER COLUMN lesson_uid SET NOT NULL;
-- foreign key: absents lesson_uid
ALTER TABLE absents ADD COLUMN lesson_uid uuid;
UPDATE absents tt SET lesson_uid = c.uid FROM subjects c WHERE c.id = tt.lesson_id;
ALTER TABLE absents DROP COLUMN lesson_id;
ALTER TABLE absents ADD FOREIGN KEY (lesson_uid)  REFERENCES lessons(uid) ON DELETE SET NULL;
-- ALTER TABLE absents ALTER COLUMN lesson_uid SET NOT NULL;
-- foreign key: lesson_likes lesson_uid
ALTER TABLE lesson_likes ADD COLUMN lesson_uid uuid;
UPDATE lesson_likes tt SET lesson_uid = c.uid FROM subjects c WHERE c.id = tt.lesson_id;
ALTER TABLE lesson_likes DROP COLUMN lesson_id;
ALTER TABLE lesson_likes ADD FOREIGN KEY (lesson_uid)  REFERENCES lessons(uid) ON DELETE SET NULL;
-- ALTER TABLE lesson_likes ALTER COLUMN lesson_uid SET NOT NULL;

ALTER TABLE sessions ADD COLUMN user_uid uuid;
UPDATE sessions tt SET user_uid = c.uid FROM users c WHERE c.id = tt.user_id;
ALTER TABLE sessions DROP COLUMN user_id;
ALTER TABLE sessions ADD FOREIGN KEY (user_uid)  REFERENCES users(uid) ON DELETE SET NULL;

-- foreign key: messages group_uid
ALTER TABLE messages ADD COLUMN group_uid uuid;
UPDATE messages tt SET group_uid = c.uid FROM message_groups c WHERE c.id = tt.group_id;

ALTER TABLE messages DROP COLUMN group_id;
ALTER TABLE messages ADD FOREIGN KEY (group_uid)  REFERENCES message_groups(uid) ON DELETE SET NULL;
-- ALTER TABLE messages ALTER COLUMN group_uid SET NOT NULL;


-- foreign key: messages parent_uid
ALTER TABLE messages ADD COLUMN parent_uid uuid;
UPDATE messages tt SET parent_uid = c.uid FROM messages c WHERE c.id = tt.parent_id;

ALTER TABLE messages DROP COLUMN parent_id;
ALTER TABLE messages ADD FOREIGN KEY (parent_uid)  REFERENCES messages(uid) ON DELETE SET NULL;
-- ALTER TABLE messages ALTER COLUMN parent_uid SET NOT NULL;
-- foreign key: messages_reads message_uid
ALTER TABLE messages_reads ADD COLUMN message_uid uuid;
UPDATE messages_reads tt SET message_uid = c.uid FROM messages c WHERE c.id = tt.message_id;

ALTER TABLE messages_reads DROP COLUMN message_id;
ALTER TABLE messages_reads ADD FOREIGN KEY (message_uid)  REFERENCES messages(uid) ON DELETE SET NULL;
-- ALTER TABLE messages_reads ALTER COLUMN message_uid SET NOT NULL;

ALTER TABLE messages_reads ADD COLUMN session_uid uuid;
UPDATE messages_reads tt SET session_uid = c.uid FROM sessions c WHERE c.id = tt.session_id;

ALTER TABLE messages_reads DROP COLUMN session_id;
ALTER TABLE messages_reads ADD FOREIGN KEY (session_uid)  REFERENCES sessions(uid) ON DELETE SET NULL;
-- ALTER TABLE messages_reads ALTER COLUMN message_uid SET NOT NULL;


-- foreign key: user_notifications notification_uid
ALTER TABLE user_notifications ADD COLUMN notification_uid uuid;
UPDATE user_notifications tt SET notification_uid = c.uid FROM notifications c WHERE c.id = tt.notification_id;

ALTER TABLE user_notifications DROP COLUMN notification_id;
ALTER TABLE user_notifications ADD FOREIGN KEY (notification_uid)  REFERENCES notifications(uid) ON DELETE SET NULL;
-- ALTER TABLE user_notifications ALTER COLUMN notification_uid SET NOT NULL;

ALTER TABLE reports ADD COLUMN school_uids uuid[];
UPDATE reports tt SET school_uids = s.uid FROM schools s WHERE s.id = ANY(tt.school_ids);
ALTER TABLE reports DROP COLUMN school_ids;

-- foreign key: report_items report_uid
ALTER TABLE report_items ADD COLUMN report_uid uuid;
UPDATE report_items tt SET report_uid = c.uid FROM reports c WHERE c.id = tt.report_id;

ALTER TABLE report_items DROP COLUMN report_id;
ALTER TABLE report_items ADD FOREIGN KEY (report_uid)  REFERENCES reports(uid) ON DELETE SET NULL;
-- ALTER TABLE report_items ALTER COLUMN report_uid SET NOT NULL;

ALTER TABLE confirm_codes ADD COLUMN user_uid uuid;
UPDATE confirm_codes cc SET user_uid = u.uid FROM users u WHERE u.id = cc.user_id;
ALTER TABLE confirm_codes DROP COLUMN user_id;
ALTER TABLE confirm_codes ADD FOREIGN KEY (user_uid) REFERENCES users(uid) ON DELETE CASCADE; 

ALTER TABLE topics ADD COLUMN book_uid uuid;
UPDATE topics cc SET book_uid = b.uid FROM books b WHERE b.id = cc.book_id;
ALTER TABLE topics DROP COLUMN book_id;
ALTER TABLE topics ADD FOREIGN KEY (book_uid) REFERENCES books(uid) ON DELETE CASCADE; 

ALTER TABLE assignments ADD COLUMN created_by_uid uuid;
UPDATE assignments cc SET created_by_uid = b.uid FROM users b WHERE b.id = cc.created_by;
ALTER TABLE assignments DROP COLUMN created_by;
ALTER TABLE assignments ADD FOREIGN KEY (created_by_uid) REFERENCES users(uid) ON DELETE CASCADE; 

ALTER TABLE assignments ADD COLUMN updated_by_uid uuid;
UPDATE assignments cc SET updated_by_uid = b.uid FROM users b WHERE b.id = cc.updated_by;
ALTER TABLE assignments DROP COLUMN updated_by;
ALTER TABLE assignments ADD FOREIGN KEY (updated_by_uid) REFERENCES users(uid) ON DELETE CASCADE; 

ALTER TABLE assignments ADD COLUMN lesson_uid uuid;
UPDATE assignments cc SET lesson_uid = l.uid FROM lessons l WHERE l.id = cc.lesson_id;
ALTER TABLE assignments DROP COLUMN lesson_id;
ALTER TABLE assignments ADD FOREIGN KEY (lesson_uid) REFERENCES lessons(uid) ON DELETE CASCADE; 


ALTER TABLE users DROP COLUMN id;
ALTER TABLE users ADD PRIMARY KEY (uid);
ALTER TABLE schools DROP COLUMN id;
ALTER TABLE schools ADD PRIMARY KEY (uid);
ALTER TABLE classrooms DROP COLUMN id;
ALTER TABLE classrooms ADD PRIMARY KEY (uid);
ALTER TABLE subjects DROP COLUMN id;
ALTER TABLE subjects ADD PRIMARY KEY (uid);
ALTER TABLE periods DROP COLUMN id;
ALTER TABLE periods ADD PRIMARY KEY (uid);
ALTER TABLE shifts DROP COLUMN id;
ALTER TABLE shifts ADD PRIMARY KEY (uid);
ALTER TABLE base_subjects DROP COLUMN id;
ALTER TABLE base_subjects ADD PRIMARY KEY (uid);
ALTER TABLE lessons DROP COLUMN id;
ALTER TABLE lessons ADD PRIMARY KEY (uid);
ALTER TABLE timetables DROP COLUMN id;
ALTER TABLE timetables ADD PRIMARY KEY (uid);
ALTER TABLE grades DROP COLUMN id;
ALTER TABLE grades ADD PRIMARY KEY (uid);
ALTER TABLE absents DROP COLUMN id;
ALTER TABLE absents ADD PRIMARY KEY (uid);
ALTER TABLE period_grades DROP COLUMN id;
ALTER TABLE period_grades ADD PRIMARY KEY (uid);
ALTER TABLE user_logs DROP COLUMN id;
ALTER TABLE user_logs ADD PRIMARY KEY (uid);
ALTER TABLE payment_transactions DROP COLUMN id;
ALTER TABLE payment_transactions ADD PRIMARY KEY (uid);
ALTER TABLE message_groups DROP COLUMN id;
ALTER TABLE message_groups ADD PRIMARY KEY (uid);
ALTER TABLE messages DROP COLUMN id;
ALTER TABLE messages ADD PRIMARY KEY (uid);
ALTER TABLE messages_reads DROP COLUMN id;
ALTER TABLE messages_reads ADD PRIMARY KEY (uid);
ALTER TABLE notifications DROP COLUMN id;
ALTER TABLE notifications ADD PRIMARY KEY (uid);
ALTER TABLE user_notifications DROP COLUMN id;
ALTER TABLE user_notifications ADD PRIMARY KEY (uid);
ALTER TABLE reports DROP COLUMN id;
ALTER TABLE reports ADD PRIMARY KEY (uid);
ALTER TABLE subject_exams DROP COLUMN id;
ALTER TABLE subject_exams ADD PRIMARY KEY (uid);
ALTER TABLE confirm_codes DROP COLUMN id;
ALTER TABLE confirm_codes ADD PRIMARY KEY (uid);
ALTER TABLE report_items DROP COLUMN id;
ALTER TABLE report_items ADD PRIMARY KEY (uid);
ALTER TABLE user_schools DROP COLUMN id;
ALTER TABLE user_schools ADD PRIMARY KEY (uid);
ALTER TABLE user_classrooms DROP COLUMN id;
ALTER TABLE user_classrooms ADD PRIMARY KEY (uid);
ALTER TABLE user_parents DROP COLUMN id;
ALTER TABLE user_parents ADD PRIMARY KEY (uid);
ALTER TABLE user_settings DROP COLUMN id;
ALTER TABLE user_settings ADD PRIMARY KEY (uid);
ALTER TABLE student_notes DROP COLUMN id;
ALTER TABLE student_notes ADD PRIMARY KEY (uid);
ALTER TABLE school_settings DROP COLUMN id;
ALTER TABLE school_settings ADD PRIMARY KEY (uid);
ALTER TABLE sessions DROP COLUMN id;
ALTER TABLE sessions ADD PRIMARY KEY (uid);
ALTER TABLE books DROP COLUMN id;
ALTER TABLE books ADD PRIMARY KEY (uid);
ALTER TABLE topics DROP COLUMN id;
ALTER TABLE topics ADD PRIMARY KEY (uid);
ALTER TABLE sms_sender DROP COLUMN id;
ALTER TABLE sms_sender ADD PRIMARY KEY (uid);
ALTER TABLE assignments DROP COLUMN id;
ALTER TABLE assignments ADD PRIMARY KEY (uid);
ALTER TABLE contact_items DROP COLUMN id;
ALTER TABLE contact_items ADD PRIMARY KEY (uid);
ALTER TABLE lesson_likes DROP COLUMN id;
ALTER TABLE lesson_likes ADD PRIMARY KEY (uid);