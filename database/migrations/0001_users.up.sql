CREATE TABLE users (
  id serial primary KEY,
  first_name varchar(255) NOT NULL,
  middle_name varchar(255) DEFAULT NULL,
  last_name varchar(255) DEFAULT NULL,
  username varchar(255) NOT NULL,
  password varchar(255) NOT NULL,
  status varchar(255) NOT NULL,
  phone varchar(255) DEFAULT NULL,
  phone_verified_at timestamp NULL DEFAULT NULL,
  email varchar(255) DEFAULT NULL,
  email_verified_at timestamp NULL DEFAULT NULL,
  birthday date DEFAULT NULL,
  gender integer DEFAULT NULL,
  address varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  last_active timestamp DEFAULT NULL,
  passport_number varchar(255) DEFAULT NULL,
  birth_cert_number varchar(255) DEFAULT NULL,
  apply_number varchar(255) DEFAULT NULL,
  work_title varchar(255) DEFAULT NULL,
  work_place varchar(255) DEFAULT NULL,
  district varchar(255) DEFAULT NULL,
  reference varchar(255) DEFAULT NULL,
  nickname varchar(255) DEFAULT NULL,
  education_title varchar(255) DEFAULT NULL,
  education_place varchar(255) DEFAULT NULL,
  education_group varchar(255) DEFAULT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  archived_at timestamp DEFAULT NULL,
  last_paid_at timestamp DEFAULT NULL,
  documents json DEFAULT NULL,
  document_files varchar(255)[] DEFAULT NULL,
  test1 varchar(255) DEFAULT NULL,
  test2 varchar(255) DEFAULT NULL,
  test3 varchar(255) DEFAULT NULL,
  test4 varchar(255) DEFAULT NULL,
  test5 varchar(255) DEFAULT NULL,
  test6 varchar(255) DEFAULT NULL,
  test7 varchar(255) DEFAULT NULL,
  test8 varchar(255) DEFAULT NULL,
  test9 varchar(255) DEFAULT NULL,
  test10 varchar(255) DEFAULT NULL,
  test11 varchar(255) DEFAULT NULL,
  test12 varchar(255) DEFAULT NULL,
  test13 varchar(255) DEFAULT NULL,
  test14 varchar(255) DEFAULT NULL,
  test15 varchar(255) DEFAULT NULL,
  test16 varchar(255) DEFAULT NULL,
  test17 varchar(255) DEFAULT NULL,
  test18 varchar(255) DEFAULT NULL,
  test19 varchar(255) DEFAULT NULL,
  test20 varchar(255) DEFAULT NULL,
  test21 varchar(255) DEFAULT NULL,
  test22 varchar(255) DEFAULT NULL,
  test23 varchar(255) DEFAULT NULL,
  test24 varchar(255) DEFAULT NULL,
  test25 varchar(255) DEFAULT NULL,
  test26 varchar(255) DEFAULT NULL,
  -- UNIQUE KEY users_username_unique (username),
  -- KEY users_username_index (username),
  -- KEY users_phone_index (phone),
  -- KEY users_email_index (email),
  -- KEY users_gender_index (gender),
  -- KEY users_status_index (status)
  UNIQUE(username)
);




CREATE TABLE user_settings (
  id serial primary KEY ,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  key varchar(255) NOT NULL,
  value text DEFAULT NULL
);



CREATE TABLE user_parents (
  id serial primary KEY,
  parent_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  child_id bigint NOT NULL REFERENCES users ON DELETE CASCADE
);


CREATE TABLE notifications (
  id serial primary KEY ,
  school_ids integer[] NOT NULL,
  user_ids integer[] DEFAULT NULL,
  roles text[] DEFAULT NULL,
  author_id bigint NULL REFERENCES users,
  title varchar(255) NOT NULL,
  content text DEFAULT NULL,
  files text[] DEFAULT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE user_notifications (
  id serial primary KEY,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  notification_id bigint NOT NULL REFERENCES notifications ON DELETE CASCADE,
  role varchar(255) DEFAULT NULL,
  read_at timestamp DEFAULT NULL,
  comment text DEFAULT NULL,
  comment_files text[] DEFAULT NULL
);


CREATE TABLE confirm_codes (
  id serial primary KEY ,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  phone varchar(255) DEFAULT NULL,
  email varchar(255) DEFAULT NULL,
  code varchar(255) NOT NULL,
  expire_at timestamp NOT NULL
);


CREATE TABLE sessions (
  id serial primary KEY,
  token varchar(255) NOT NULL,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  device_token text DEFAULT NULL,
  agent text NOT NULL,
  ip varchar(255) NOT NULL,
  iat timestamp DEFAULT NULL,
  exp timestamp DEFAULT NULL,
  lat timestamp DEFAULT NULL
);


CREATE TABLE user_logs (
  id serial primary KEY,
  school_id bigint DEFAULT NULL,
  user_id bigint NOT NULL,
  session_id bigint NOT NULL,
  subject_id bigint DEFAULT NULL,
  subject varchar(255) DEFAULT NULL,
  subject_action varchar(255) NOT NULL,
  subject_description varchar(512) DEFAULT NULL,
  properties json DEFAULT null,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);
