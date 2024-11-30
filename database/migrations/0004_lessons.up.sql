CREATE TABLE books (
  id serial primary key,
  title varchar(255) DEFAULT NULL,
  categories varchar(255)[] DEFAULT NULL,
  description text DEFAULT NULL,
  year int DEFAULT NULL,
  pages int DEFAULT NULL,
  authors varchar(255)[] DEFAULT NULL,
  file varchar(255) DEFAULT NULL,
  file_size int DEFAULT NULL,
  file_preview varchar(255) DEFAULT NULL,
  is_downloadable boolean,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE lessons (
  id serial primary KEY ,
  school_id bigint NOT NULL REFERENCES schools ON DELETE CASCADE,
  subject_id bigint NOT NULL REFERENCES subjects ON DELETE CASCADE,
  book_id bigint DEFAULT NULL REFERENCES books ON DELETE SET NULL,
  book_page bigint DEFAULT NULL,
  period_id bigint DEFAULT NULL REFERENCES periods ON DELETE SET NULL,
  period_key bigint DEFAULT 0,
  "date" date NOT NULL,
  hour_number integer DEFAULT NULL,
  type_title varchar(255) DEFAULT NULL,
  title varchar(255) DEFAULT NULL,
  content text DEFAULT NULL,
  pro_title varchar(255) DEFAULT NULL,
  pro_files json DEFAULT NULL,
  assignment_title varchar(512) DEFAULT NULL,
  assignment_content text DEFAULT NULL,
  assignment_files json DEFAULT NULL,
  lesson_attributes json DEFAULT NULL,
  is_teacher_excused boolean DEFAULT false,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE lesson_likes (
  id serial primary key,
  user_id bigint DEFAULT NULL REFERENCES users ON DELETE SET NULL,
  lesson_id bigint DEFAULT NULL REFERENCES lessons ON DELETE SET NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE grades (
  id serial primary KEY,
  lesson_id bigint NOT NULL REFERENCES lessons ON DELETE CASCADE,
  student_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  "value" varchar(255) DEFAULT NULL,
  "values" integer[] DEFAULT NULL,
  comment text DEFAULT NULL,
  reason varchar(255) DEFAULT NULL,
  parent_reviewed_at timestamp DEFAULT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  deleted_at timestamp DEFAULT NULL,
  created_by bigint NOT NULL REFERENCES users,
  updated_by bigint NOT NULL REFERENCES users
);


CREATE TABLE absents (
  id serial primary KEY,
  lesson_id bigint NOT NULL REFERENCES lessons ON DELETE CASCADE,
  student_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  comment text DEFAULT NULL,
  reason varchar(255) NOT NULL,
  parent_reviewed_at timestamp DEFAULT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  deleted_at timestamp DEFAULT NULL,
  created_by bigint NOT NULL REFERENCES users,
  updated_by bigint NOT NULL REFERENCES users
);


CREATE TABLE assignments (
  id serial primary KEY ,
  lesson_id bigint NOT NULL REFERENCES lessons ON DELETE CASCADE,
  title varchar(512) NOT NULL,
  content text DEFAULT NULL,
  files json DEFAULT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  created_by bigint NOT NULL REFERENCES users,
  updated_by bigint NOT NULL REFERENCES users
);


CREATE TABLE topics (
  id serial primary KEY,
  book_id bigint DEFAULT NULL REFERENCES books ON DELETE SET NULL,
  book_page bigint DEFAULT NULL,
  subject_name varchar(255) NOT NULL,
  classyear varchar(255) DEFAULT NULL,
  period varchar(255) DEFAULT NULL,
  level varchar(255) DEFAULT NULL,
  language varchar(255) DEFAULT NULL,
  tags text[] DEFAULT NULL,
  title varchar(255) NOT NULL,
  content varchar(255) DEFAULT NULL,
  files text[] DEFAULT NULL
);
