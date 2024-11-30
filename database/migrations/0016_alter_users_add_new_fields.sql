ALTER TABLE users ADD passport_number varchar(255) DEFAULT NULL;
ALTER TABLE users ADD birth_cert_number varchar(255) DEFAULT NULL;
ALTER TABLE users ADD work_title varchar(255) DEFAULT NULL;
ALTER TABLE users ADD work_place varchar(255) DEFAULT NULL;
ALTER TABLE users ADD district varchar(255) DEFAULT NULL;
ALTER TABLE users ADD reference varchar(255) DEFAULT NULL;
ALTER TABLE users ADD nickname varchar(255) DEFAULT NULL;
ALTER TABLE users ADD education_title varchar(255) DEFAULT NULL;
ALTER TABLE users ADD education_place varchar(255) DEFAULT NULL;
ALTER TABLE users ADD education_group varchar(255) DEFAULT NULL;
ALTER TABLE users ADD apply_number varchar(255) DEFAULT NULL;

ALTER TABLE user_classrooms ADD COLUMN tariff_type varchar(255) DEFAULT NULL;
ALTER TABLE user_classrooms ADD COLUMN tariff_end_at timestamp DEFAULT NULL;
