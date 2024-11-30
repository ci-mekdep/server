CREATE TABLE settings (
  uid uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  key varchar(255) NOT NULL,
  value text DEFAULT NULL,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (key)
);