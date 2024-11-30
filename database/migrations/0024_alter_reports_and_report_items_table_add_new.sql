ALTER TABLE report_items ADD COLUMN classroom_uid uuid;
ALTER TABLE report_items ADD COLUMN updated_by uuid;
ALTER TABLE report_items ADD COLUMN is_edited_manually boolean;

ALTER TABLE reports ADD COLUMN is_classrooms_included boolean DEFAULT false;
ALTER TABLE reports ADD COLUMN region_uids uuid[];