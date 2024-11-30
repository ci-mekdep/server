ALTER TABLE schools ADD galleries varchar(255)[] DEFAULT NULL;
ALTER TABLE schools ADD latitude varchar(255) DEFAULT NULL;
ALTER TABLE schools ADD longitude varchar(255) DEFAULT NULL;
ALTER TABLE schools ADD is_digitalized boolean;
ALTER TABLE schools ADD is_secondary_school boolean DEFAULT TRUE;