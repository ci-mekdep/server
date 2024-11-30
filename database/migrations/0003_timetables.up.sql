CREATE TABLE base_subjects (
  id serial primary KEY,
  school_id bigint DEFAULT NULL REFERENCES schools ON DELETE SET NULL,
  name varchar(255) DEFAULT NULL,
  category varchar(255) DEFAULT NULL,
  price bigint DEFAULT NULL,
  exam_min_grade bigint DEFAULT NULL,
  age_category varchar(255) DEFAULT NULL,
  is_available boolean,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE  subjects (
  id serial primary KEY ,
  school_id bigint NOT NULL REFERENCES schools ON DELETE CASCADE,
  classroom_id bigint NOT NULL REFERENCES classrooms ON DELETE CASCADE,
  classroom_type varchar(255) DEFAULT NULL,
  classroom_type_key int DEFAULT NULL,
  period_id bigint DEFAULT NULL REFERENCES periods ON DELETE SET NULL,
  base_subject_id bigint DEFAULT NULL REFERENCES base_subjects ON DELETE SET NULL,
  parent_id bigint DEFAULT NULL REFERENCES subjects ON DELETE SET NULL,
  name varchar(255) NOT NULL,
  full_name varchar(255) DEFAULT NULL,
  teacher_id bigint DEFAULT NULL REFERENCES users ON DELETE SET NULL,
  second_teacher_id bigint DEFAULT NULL REFERENCES users ON DELETE SET NULL,
  week_hours int DEFAULT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE subject_exams (
  id serial primary KEY,
  subject_id bigint NOT NULL REFERENCES subjects,
  school_id bigint NOT NULL REFERENCES schools,
  classroom_id bigint NOT NULL REFERENCES classrooms,
  teacher_id bigint NOT NULL REFERENCES users,
  head_teacher_id bigint DEFAULT NULL REFERENCES users,
  member_teacher_ids integer[] DEFAULT NULL,
  room_number varchar(255) DEFAULT NULL,
  start_time timestamp DEFAULT CURRENT_TIMESTAMP,
  time_length_min integer DEFAULT NULL,
  exam_weight_percent bigint DEFAULT NULL,
  name varchar(255) DEFAULT NULL,
  is_required boolean,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP  
);


CREATE TABLE period_grades (
  id serial primary KEY ,
  period_id bigint NOT NULL REFERENCES periods ON DELETE CASCADE,
  period_key int NOT NULL DEFAULT 0,
  subject_id bigint NOT NULL,
  student_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  exam_id bigint DEFAULT NULL REFERENCES subject_exams ON DELETE CASCADE,
  lesson_count int NOT NULL default 0,
  absent_count int NOT NULL default 0,
  grade_count int NOT NULL default 0,
  grade_sum int NOT NULL default 0,
  prev_grade_count int NOT NULL default 0,
  prev_grade_sum int NOT NULL default 0,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE  timetables (
  id serial primary KEY ,
  school_id bigint NOT NULL REFERENCES schools ON DELETE CASCADE,
  classroom_id bigint NOT NULL REFERENCES classrooms ON DELETE CASCADE,
  shift_id bigint NOT NULL REFERENCES shifts ON DELETE CASCADE,
  period_id bigint DEFAULT NULL REFERENCES periods ON DELETE CASCADE,
  value text DEFAULT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_by bigint DEFAULT NULL REFERENCES users
);


CREATE TABLE  student_notes (
  id serial primary KEY ,
  school_id bigint NOT NULL REFERENCES schools ON DELETE CASCADE,
  subject_id bigint DEFAULT NULL REFERENCES subjects ON DELETE CASCADE,
  student_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  teacher_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  note text DEFAULT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);
