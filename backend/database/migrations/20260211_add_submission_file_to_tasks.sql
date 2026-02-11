-- Add submission_file column to tasks table
ALTER TABLE tasks ADD COLUMN submission_file varchar(255) DEFAULT NULL;
