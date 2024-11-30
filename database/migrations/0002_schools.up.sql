CREATE TABLE schools (
  id serial primary KEY ,
  code varchar(255) NOT NULL,
  name varchar(255) NOT NULL,
  full_name varchar(255) DEFAULT NULL,
  description text DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  background varchar(255) DEFAULT NULL,
  phone varchar(255) DEFAULT NULL,
  email varchar(255) DEFAULT NULL,
  address varchar(255) DEFAULT NULL,
  level varchar(255) DEFAULT NULL,
  galleries varchar(255)[] DEFAULT NULL,
  latitude varchar(255) DEFAULT NULL,
  longitude varchar(255) DEFAULT NULL,
  is_digitalized boolean,
  is_secondary_school boolean DEFAULT TRUE,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  archived_at timestamp DEFAULT NULL,
  parent_id bigint DEFAULT NULL REFERENCES schools ON DELETE SET NULL,
  admin_id bigint DEFAULT NULL REFERENCES users ON DELETE SET NULL,
  specialist_id bigint DEFAULT NULL REFERENCES users ON DELETE SET NULL,
  test1 varchar(255) DEFAULT NULL,
  UNIQUE(code)
);


CREATE TABLE user_schools (
  id serial primary KEY ,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  school_id bigint DEFAULT NULL REFERENCES schools ON DELETE CASCADE,
  role_code varchar(255) NOT NULL
);


CREATE TABLE school_settings (
  id serial primary KEY ,
  school_id bigint NOT NULL REFERENCES schools ON DELETE CASCADE,
  key varchar(255) NOT NULL,
  value text DEFAULT NULL,
  UNIQUE (key, school_id)
);


CREATE TABLE  shifts (
  id serial primary KEY,
  school_id bigint NOT NULL REFERENCES schools ON DELETE CASCADE,
  name varchar(255) NOT NULL,
  value text DEFAULT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_by bigint DEFAULT NULL REFERENCES users
);

CREATE TABLE classrooms (
  id serial primary KEY ,
  school_id bigint NOT NULL REFERENCES schools ON DELETE CASCADE,
  shift_id bigint DEFAULT NULL REFERENCES shifts ON DELETE SET NULL,
  name varchar(255) NOT NULL,
  name_canonical varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  description text DEFAULT NULL,
  language varchar(255) DEFAULT NULL,
  level varchar(255) DEFAULT NULL,
  teacher_id bigint DEFAULT NULL REFERENCES users ON DELETE SET NULL,
  student_id bigint DEFAULT NULL REFERENCES users ON DELETE SET NULL,
  parent_id bigint DEFAULT NULL REFERENCES classrooms ON DELETE SET NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  archived_at timestamp DEFAULT NULL,
  test1 varchar(255) DEFAULT NULL
);


CREATE TABLE user_classrooms (
  id serial primary KEY ,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  classroom_id bigint NOT NULL REFERENCES classrooms ON DELETE CASCADE,
  type varchar(255) DEFAULT NULL,
  type_key int DEFAULT NULL,
  tariff_end_at timestamp DEFAULT NULL,
  tariff_type varchar(255) DEFAULT NULL
);


CREATE TABLE periods (
  id serial primary KEY ,
  school_id bigint NOT NULL REFERENCES schools ON DELETE CASCADE,
  title varchar(255) NOT NULL,
  value json default null,
  data_counts json default null,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);
