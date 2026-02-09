USE interna_db;

-- USERS
ALTER TABLE users ADD COLUMN name VARCHAR(255);
ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMP NULL;
ALTER TABLE users ADD COLUMN avatar VARCHAR(255);
ALTER TABLE users ADD COLUMN google_id VARCHAR(255);
ALTER TABLE users ADD COLUMN provider VARCHAR(50);
ALTER TABLE users ADD COLUMN google2fa_secret VARCHAR(255);
ALTER TABLE users ADD COLUMN google2fa_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN two_factor_secret TEXT;
ALTER TABLE users ADD COLUMN two_factor_recovery_codes TEXT;
ALTER TABLE users ADD COLUMN two_factor_confirmed_at TIMESTAMP NULL;
ALTER TABLE users ADD COLUMN remember_token VARCHAR(100);
ALTER TABLE users MODIFY COLUMN password_hash VARCHAR(255) NULL;
ALTER TABLE users MODIFY COLUMN role ENUM('admin','supervisor','pembimbing','intern') NOT NULL;

-- SUPERVISORS
ALTER TABLE supervisors ADD COLUMN address TEXT;
ALTER TABLE supervisors ADD COLUMN institution VARCHAR(255);
ALTER TABLE supervisors ADD COLUMN status ENUM('pending','active') DEFAULT 'active';

-- INTERNS
ALTER TABLE interns ADD COLUMN nis VARCHAR(50);
ALTER TABLE interns ADD COLUMN school VARCHAR(255);
ALTER TABLE interns ADD COLUMN department VARCHAR(255);
ALTER TABLE interns ADD COLUMN certificate_number VARCHAR(100);
ALTER TABLE interns ADD COLUMN certificate_issued_at DATE;
ALTER TABLE interns MODIFY COLUMN status ENUM('pending','active','completed','cancelled','terminated') DEFAULT 'active';
ALTER TABLE interns MODIFY COLUMN supervisor_id BIGINT NULL;

-- Update supervisor_id to reference users
ALTER TABLE interns DROP FOREIGN KEY interns_ibfk_3;
UPDATE interns i LEFT JOIN supervisors s ON i.supervisor_id = s.id
  SET i.supervisor_id = COALESCE(s.user_id, i.supervisor_id);
UPDATE interns i LEFT JOIN users u ON i.supervisor_id = u.id
  SET i.supervisor_id = NULL
  WHERE u.id IS NULL;
ALTER TABLE interns ADD CONSTRAINT interns_ibfk_3 FOREIGN KEY (supervisor_id) REFERENCES users(id) ON DELETE SET NULL;

-- TASKS
ALTER TABLE tasks ADD COLUMN task_assignment_id BIGINT;
ALTER TABLE tasks ADD COLUMN priority ENUM('low','medium','high') DEFAULT 'medium';
ALTER TABLE tasks ADD COLUMN deadline DATE;
ALTER TABLE tasks ADD COLUMN deadline_time TIME;
ALTER TABLE tasks ADD COLUMN started_at DATETIME;
ALTER TABLE tasks ADD COLUMN submitted_at DATETIME;
ALTER TABLE tasks ADD COLUMN completed_at DATETIME;
ALTER TABLE tasks ADD COLUMN approved_at DATETIME;
ALTER TABLE tasks ADD COLUMN is_late BOOLEAN DEFAULT FALSE;
ALTER TABLE tasks ADD COLUMN submission_notes TEXT;
ALTER TABLE tasks ADD COLUMN submission_links JSON;
ALTER TABLE tasks ADD COLUMN score INT;
ALTER TABLE tasks ADD COLUMN admin_feedback TEXT;
ALTER TABLE tasks MODIFY COLUMN status ENUM('pending','scheduled','in_progress','submitted','revision','completed','overdue','cancelled') DEFAULT 'pending';
ALTER TABLE tasks MODIFY COLUMN description TEXT NULL;

-- Move assigned_by to users
ALTER TABLE tasks DROP FOREIGN KEY tasks_ibfk_2;
UPDATE tasks t LEFT JOIN supervisors s ON t.assigned_by = s.id
  SET t.assigned_by = COALESCE(s.user_id, t.assigned_by);
UPDATE tasks t LEFT JOIN users u ON t.assigned_by = u.id
  SET t.assigned_by = NULL
  WHERE u.id IS NULL;
SET @admin_id := (SELECT id FROM users WHERE role='admin' ORDER BY id LIMIT 1);
UPDATE tasks SET assigned_by = @admin_id WHERE assigned_by IS NULL AND @admin_id IS NOT NULL;
ALTER TABLE tasks MODIFY COLUMN assigned_by BIGINT NOT NULL;
ALTER TABLE tasks ADD CONSTRAINT tasks_ibfk_2 FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE CASCADE;

-- ATTENDANCES
ALTER TABLE attendances ADD COLUMN notes TEXT;
ALTER TABLE attendances ADD COLUMN distance_meters INT;
ALTER TABLE attendances ADD COLUMN proof_file VARCHAR(500);
ALTER TABLE attendances MODIFY COLUMN status ENUM('present','late','absent','sick','permission','on_leave','excused') DEFAULT 'absent';
ALTER TABLE attendances ADD UNIQUE KEY unique_intern_date (intern_id, date);

-- ASSESSMENTS
ALTER TABLE assessments ADD COLUMN quality_score INT;
ALTER TABLE assessments ADD COLUMN speed_score INT;
ALTER TABLE assessments ADD COLUMN initiative_score INT;
ALTER TABLE assessments ADD COLUMN teamwork_score INT;
ALTER TABLE assessments ADD COLUMN communication_score INT;
ALTER TABLE assessments ADD COLUMN strengths TEXT;
ALTER TABLE assessments ADD COLUMN improvements TEXT;
ALTER TABLE assessments ADD COLUMN comments TEXT;

-- NEW TABLES
CREATE TABLE IF NOT EXISTS task_assignments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    assigned_by BIGINT NOT NULL,
    priority ENUM('low', 'medium', 'high') DEFAULT 'medium',
    start_date DATE,
    deadline DATE,
    deadline_time TIME,
    assign_to_all BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS task_assignment_interns (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_assignment_id BIGINT NOT NULL,
    intern_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (task_assignment_id) REFERENCES task_assignments(id) ON DELETE CASCADE,
    FOREIGN KEY (intern_id) REFERENCES interns(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS reports (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    intern_id BIGINT NOT NULL,
    created_by BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    type ENUM('weekly', 'monthly', 'final') DEFAULT 'weekly',
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    status ENUM('draft', 'submitted', 'reviewed') DEFAULT 'draft',
    feedback TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (intern_id) REFERENCES interns(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS settings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    `key` VARCHAR(255) UNIQUE NOT NULL,
    `value` TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS notifications (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type ENUM('info', 'success', 'warning', 'error') DEFAULT 'info',
    title VARCHAR(255) NOT NULL,
    message TEXT,
    icon VARCHAR(100),
    link VARCHAR(255),
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS password_resets (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- Backfill task deadlines from legacy target_date
UPDATE tasks SET deadline = target_date WHERE deadline IS NULL;
