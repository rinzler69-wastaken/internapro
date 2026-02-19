-- Ensure newer task assignment columns exist on older production schemas.
-- Safe to run multiple times.
SET @db_name := DATABASE();

-- tasks.is_unscheduled
SET @has_is_unscheduled := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
    AND TABLE_NAME = 'tasks'
    AND COLUMN_NAME = 'is_unscheduled'
);
SET @sql := IF(
  @has_is_unscheduled = 0,
  "ALTER TABLE `tasks` ADD COLUMN `is_unscheduled` TINYINT(1) DEFAULT '0'",
  "SELECT 'skip: tasks.is_unscheduled already exists'"
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- tasks.assigner_id
SET @has_assigner_id := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
    AND TABLE_NAME = 'tasks'
    AND COLUMN_NAME = 'assigner_id'
);
SET @sql := IF(
  @has_assigner_id = 0,
  "ALTER TABLE `tasks` ADD COLUMN `assigner_id` BIGINT DEFAULT NULL",
  "SELECT 'skip: tasks.assigner_id already exists'"
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- tasks.custom_assigner_name
SET @has_custom_assigner_name := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
    AND TABLE_NAME = 'tasks'
    AND COLUMN_NAME = 'custom_assigner_name'
);
SET @sql := IF(
  @has_custom_assigner_name = 0,
  "ALTER TABLE `tasks` ADD COLUMN `custom_assigner_name` VARCHAR(255) DEFAULT NULL",
  "SELECT 'skip: tasks.custom_assigner_name already exists'"
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- index for assigner_id
SET @has_assigner_index := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = @db_name
    AND TABLE_NAME = 'tasks'
    AND INDEX_NAME = 'fk_tasks_assigner'
);
SET @sql := IF(
  @has_assigner_index = 0,
  "ALTER TABLE `tasks` ADD KEY `fk_tasks_assigner` (`assigner_id`)",
  "SELECT 'skip: index fk_tasks_assigner already exists'"
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- foreign key tasks.assigner_id -> users.id
SET @has_assigner_fk := (
  SELECT COUNT(*)
  FROM information_schema.TABLE_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = @db_name
    AND TABLE_NAME = 'tasks'
    AND CONSTRAINT_TYPE = 'FOREIGN KEY'
    AND CONSTRAINT_NAME = 'fk_tasks_assigner'
);
SET @sql := IF(
  @has_assigner_fk = 0,
  "ALTER TABLE `tasks` ADD CONSTRAINT `fk_tasks_assigner` FOREIGN KEY (`assigner_id`) REFERENCES `users`(`id`) ON DELETE SET NULL",
  "SELECT 'skip: fk_tasks_assigner already exists'"
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
