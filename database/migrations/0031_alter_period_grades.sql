ALTER TABLE period_grades
ADD COLUMN old_absent_count int NOT NULL DEFAULT 0,
ADD COLUMN old_grade_count int NOT NULL DEFAULT 0,
ADD COLUMN old_grade_sum int NOT NULL DEFAULT 0;