CREATE TABLE payment_transactions (
  id serial primary KEY ,
  school_id bigint NOT NULL REFERENCES schools ON DELETE CASCADE,
  payer_id bigint NOT NULL REFERENCES users,
  user_ids bigint[] NOT NULL,
  tariff_type varchar(255) NOT NULL,
  status varchar(255) NOT NULL,
  amount float NOT NULL,
  original_amount float NOT NULL,
  bank_type varchar(255) NOT NULL,
  card_name varchar(255) DEFAULT NULL,
  order_number varchar(255) DEFAULT NULL,
  order_url varchar(255) DEFAULT NULL,
  comment text DEFAULT NULL,
  system_comment text DEFAULT NULL,
  unit_price float DEFAULT NULL,
  school_price float DEFAULT NULL,
  center_price float DEFAULT NULL,
  discount_price float DEFAULT NULL,
  used_days varchar(255) DEFAULT NULL,
  used_days_price float DEFAULT NULL,
  school_months int DEFAULT 0,
  center_months int DEFAULT 0,
  school_classroom_uids uuid[],
  center_classroom_uids uuid[],
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE sms_sender (
  id serial primary KEY,
  phones varchar[] NOT NULL,
  message text DEFAULT NULL,
  type varchar(255) DEFAULT NULL,
  is_completed boolean,
  left_try bigint DEFAULT NULL,
  error_msg varchar(255) DEFAULT NULL,
  tried_at timestamp DEFAULT CURRENT_TIMESTAMP,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP 
);


CREATE TABLE contact_items (
  id serial primary KEY,
  related_ids bigint[] DEFAULT NULL,
  user_id bigint DEFAULT NULL REFERENCES users ON DELETE SET NULL,
  school_id bigint DEFAULT NULL REFERENCES schools ON DELETE SET NULL,
  message text NOT NULL,
  type varchar(255) NOT NULL,
  status varchar(255) NOT NULL,
  files json DEFAULT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_by uuid DEFAULT NULL
);
