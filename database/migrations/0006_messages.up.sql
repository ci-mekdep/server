CREATE TABLE message_groups (
  id serial primary KEY,
  admin_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  title varchar(255) NOT NULL,
  description varchar(255) DEFAULT NULL,
  school_id bigint NOT NULL REFERENCES schools ON DELETE CASCADE,
  classroom_id bigint DEFAULT NULL REFERENCES classrooms ON DELETE SET NULL,
  type varchar(255) NOT NULL
);
ALTER TABLE message_groups ADD CONSTRAINT "message_groups_type" UNIQUE (type, school_id, classroom_id);


CREATE TABLE messages (
  id serial primary KEY,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  session_id bigint NOT NULL REFERENCES sessions ON DELETE CASCADE,
  group_id bigint NOT NULL REFERENCES message_groups ON DELETE CASCADE,
  parent_id bigint DEFAULT NULL REFERENCES messages ON DELETE SET NULL,
  message text DEFAULT NULL,
  files json DEFAULT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE messages_reads (
  id serial primary KEY,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  message_id bigint NOT NULL REFERENCES messages ON DELETE CASCADE,
  session_id bigint NOT NULL REFERENCES sessions ON DELETE CASCADE,
  read_at timestamp DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE messages_reads ADD CONSTRAINT "messages_reads__unique__user_id__message_id" UNIQUE (user_id, message_id);
