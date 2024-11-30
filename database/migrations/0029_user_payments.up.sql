
CREATE TABLE user_payments (
  user_uid uuid NOT NULL REFERENCES users ON DELETE CASCADE,
  classroom_uid uuid NOT NULL REFERENCES classrooms ON DELETE SET NULL,
  expires_at timestamp DEFAULT NULL
);

ALTER TABLE user_payments ADD CONSTRAINT user_payments_unique UNIQUE (user_uid, classroom_uid);
