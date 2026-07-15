ALTER TABLE courses ADD COLUMN confidence_threshold NUMERIC(3,2) NOT NULL DEFAULT 0.75;

ALTER TABLE courses ADD CONSTRAINT confidence_threshold_range
    CHECK (confidence_threshold >= 0 AND confidence_threshold <= 1);
