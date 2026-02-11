-- Add submission_method column to tasks table
ALTER TABLE tasks ADD COLUMN submission_method ENUM('links', 'files', 'both') DEFAULT 'both' AFTER description;
