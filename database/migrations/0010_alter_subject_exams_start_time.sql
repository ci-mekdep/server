ALTER TABLE subject_exams ALTER COLUMN 
    start_time SET DEFAULT NULL;
ALTER TABLE subject_exams ADD exam_weight_percent bigint DEFAULT NULL;
ALTER TABLE subject_exams ADD name varchar(255) DEFAULT NULL;
ALTER TABLE subject_exams ADD is_required boolean;
