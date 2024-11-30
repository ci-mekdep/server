CREATE TABLE reports (
    id serial primary KEY,
    title varchar(255) NOT NULL,
    description text DEFAULT NULL,
    value_types json NOT NULL,
    school_ids integer[] NOT NULL,
    is_pinned boolean DEFAULT FALSE,
    is_center_rating boolean DEFAULT FALSE,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE report_items (
    id serial primary KEY,
    report_id bigint NOT NULL REFERENCES reports ON DELETE CASCADE,
    school_id bigint DEFAULT NULL REFERENCES schools ON DELETE CASCADE,
    period_id bigint DEFAULT NULL REFERENCES periods ON DELETE CASCADE,
    values varchar[] DEFAULT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);