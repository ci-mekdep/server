create index concurrently "index_birthday_on_users" on users using btree (birthday);
create index concurrently "index_username_on_users" on users using btree (username);
create index concurrently "index_phone_on_users" on users using btree (phone);
create index concurrently "index_email_on_users" on users using btree (email);
create index concurrently "index_status_on_users" on users using btree (status);
-- create index concurrently "index_passport_number_on_users" on users using btree (passport_number);
-- create index concurrently "index_birth_cert_number_on_users" on users using btree (birth_cert_number);
-- create index concurrently "index_work_title_on_users" on users using btree (work_title);
-- create index concurrently "index_work_place_on_users" on users using btree (work_place);
-- create index concurrently "index_district_on_users" on users using btree (district);
-- create index concurrently "index_reference_on_users" on users using btree (reference);
-- create index concurrently "index_nickname_on_users" on users using btree (nickname);
-- create index concurrently "index_education_title_on_users" on users using btree (education_title);
-- create index concurrently "index_education_place_on_users" on users using btree (education_place);
-- create index concurrently "index_education_group_on_users" on users using btree (education_group);
create index concurrently "index_child_uid_on_user_users" on user_parents using btree (child_uid);
create index concurrently "index_parent_uid_on_user_users" on user_parents using btree (parent_uid);

create index concurrently "index_parent_uid_on_schools" on schools using btree (parent_uid);
create index concurrently "index_code_on_schools" on schools using btree (code);
create index concurrently "index_level_on_schools" on schools using btree (level);
-- create index concurrently "index_galleries_on_schools" on schools using btree (galleries);
-- create index concurrently "index_is_digitalized_on_schools" on schools using btree (is_digitalized);
-- create index concurrently "index_is_secondary_school_on_schools" on schools using btree (is_secondary_school);
create index concurrently "index_user_uid_on_user_schools" on user_schools using btree (user_uid);
create index concurrently "index_school_uid_on_user_schools" on user_schools using btree (school_uid);
create index concurrently "index_role_code_on_user_schools" on user_schools using btree (role_code);


create index concurrently "index_classroom_uid_on_periods" on periods using btree (school_uid);
create index concurrently "index_user_uid_on_sessions" on sessions using btree (user_uid);

create index concurrently "index_school_uid_on_classrooms" on classrooms using btree (school_uid);
create index concurrently "index_parent_uid_on_classrooms" on classrooms using btree (parent_uid);
create index concurrently "index_shift_uid_on_classrooms" on classrooms using btree (shift_uid);
create index concurrently "index_level_on_classrooms" on classrooms using btree (level);
create index concurrently "index_language_on_classrooms" on classrooms using btree (language);
create index concurrently "index_user_uid_on_user_classrooms" on user_classrooms using btree (user_uid);
create index concurrently "index_classroom_uid_on_user_classrooms" on user_classrooms using btree (classroom_uid);
create index concurrently "index_type_key_on_user_classrooms" on user_classrooms using btree (type_key);
create index concurrently "index_type_on_user_classrooms" on user_classrooms using btree (type);
create index concurrently "index_tariff_type_user_classrooms" on user_classrooms using btree (tariff_type);
create index concurrently "index_tariff_end_at_user_classrooms" on user_classrooms using btree (tariff_end_at);


create index concurrently "index_school_uid_on_timetables" on timetables using btree (school_uid);
create index concurrently "index_classroom_uid_on_timetables" on timetables using btree (classroom_uid);
create index concurrently "index_period_uid_on_timetables" on timetables using btree (period_uid);


create index concurrently "index_user_uid_on_user_notifications" on user_notifications using btree (user_uid);
create index concurrently "index_role_on_user_notifications" on user_notifications using btree (role);
create index concurrently "index_notification_uid_on_user_notifications" on user_notifications using btree (notification_uid);




create index concurrently "index_school_uid_on_subjects" on subjects using btree (school_uid);
create index concurrently "index_classroom_uid_on_subjects" on subjects using btree (classroom_uid);
create index concurrently "index_classroom_type_on_subjects" on subjects using btree (classroom_type);
create index concurrently "index_classroom_type_key_on_subjects" on subjects using btree (classroom_type_key);
create index concurrently "index_period_uid_on_subjects" on subjects using btree (period_uid);
create index concurrently "index_teacher_uid_on_subjects" on subjects using btree (teacher_uid);
create index concurrently "index_second_teacher_uid_on_subjects" on subjects using btree (second_teacher_uid);
create index concurrently "index_second_parent_uid_on_subjects" on subjects using btree (parent_uid);
create index concurrently "index_base_subject_uid_on_subjects" on subjects using btree (base_subject_uid);

create index concurrently "index_school_uid_on_base_subjects" on base_subjects using btree (school_uid);
create index concurrently "index_is_available_on_base_subjects" on base_subjects using btree (is_availabe);
-- create index concurrently "index_exam_weight_percent_on_subject_exams" on subject_exams using btree (exam_weight_percent);
-- create index concurrently "index_name_on_subject_exams" on subject_exams using btree (name);
-- create index concurrently "index_is_required_on_subject_exams" on subject_exams using btree (is_required);

create index concurrently "index_is_teacher_excused_on_lessons" on lessons using btree (is_teacher_excused);
create index concurrently "index_school_uid_on_lessons" on lessons using btree (school_uid);
create index concurrently "index_subject_uid_on_lessons" on lessons using btree (subject_uid);
create index concurrently "index_period_uid_on_lessons" on lessons using btree (period_uid);
create index concurrently "index_period_key_on_lessons" on lessons using btree (period_key);
create index concurrently "index_date_on_lessons" on lessons using btree (date);
create index concurrently "index_hour_number_on_lessons" on lessons using btree (hour_number);
create index concurrently "index_lesson_attributes_on_lessons" on lessons using btree (lesson_attributes);
create index concurrently "index_book_uid_on_lessons" on lessons using btree (book_uid);
-- create index concurrently "index_assignment_title_on_assignments" on lessons using btree (assignment_title);
-- create index concurrently "index_pro_title_on_lessons" on lessons using btree (pro_title);
-- create index concurrently "index_assignment_title_on_lessons" on lessons using btree (assignment_title);
-- create index concurrently "index_assignment_content_on_lessons" on lessons using btree (assignment_content);
-- create index concurrently "index_book_page_on_lessons" on lessons using btree (book_page);
-- create index concurrently "index_lesson_uid_on_assignments" on assignments using btree (lesson_uid);

create index concurrently "index_lesson_uid_on_grades" on grades using btree (lesson_uid);
create index concurrently "index_student_uid_on_grades" on grades using btree (student_uid);
create index concurrently "index_value_on_grades" on grades using btree (value);
create index concurrently "index_values_on_grades" on grades using btree (values);
create index concurrently "index_parent_reviewed_at_on_grades" on grades using btree (parent_reviewed_at);

create index concurrently "index_lesson_uid_on_absents" on absents using btree (lesson_uid);
create index concurrently "index_student_uid_on_absents" on absents using btree (student_uid);
create index concurrently "index_parent_reviewed_at_on_absents" on absents using btree (parent_reviewed_at);

create index concurrently "index_period_key_on_period_grades" on period_grades using btree (period_key);
create index concurrently "index_subject_uid_on_period_grades" on period_grades using btree (subject_uid);
create index concurrently "index_student_uid_on_period_grades" on period_grades using btree (student_uid);
create index concurrently "index_period_uid_on_period_grades" on period_grades using btree (period_uid);

create index concurrently "index_subject_name_on_topics" on topics using btree (subject_name);
create index concurrently "index_classyear_on_topics" on topics using btree (classyear);
create index concurrently "index_period_on_topics" on topics using btree (period);
create index concurrently "index_level_on_topics" on topics using btree (level);
create index concurrently "index_language_on_topics" on topics using btree (language);
create index concurrently "index_book_uid_on_topics" on topics using btree (book_uid);
create index concurrently "index_book_page_on_topics" on topics using btree (book_page);
create index concurrently "index_title_on_topics" on topics using btree (title);

create index concurrently "index_admin_uid_on_message_groups" on message_groups using btree (admin_uid);
create index concurrently "index_school_uid_on_message_groups" on message_groups using btree (school_uid);
create index concurrently "index_classroom_uid_on_message_groups" on message_groups using btree (classroom_uid);
create index concurrently "index_user_uid_on_messages" on messages using btree (user_uid);
create index concurrently "index_message_uid_on_messages_reads" on messages_reads using btree (message_uid);
create index concurrently "index_parent_uid_on_messages" on messages using btree (parent_uid);
-- create index concurrently "index_title_on_message_groups" on message_groups using btree (title);
-- create index concurrently "index_message_on_messages" on messages using btree (message);

create index concurrently "index_is_pinned_on_reports" on reports using btree (is_pinned);
create index concurrently "index_is_center_rating_on_reports" on reports using btree (is_center_rating);
create index concurrently "index_school_uids_on_reports" on reports using btree (school_uids);
create index concurrently "index_value_type_options_on_reports" on reports using btree (value_types_options);

create index concurrently "index_related_uid_on_contact_items" on contact_items using btree (related_uid);