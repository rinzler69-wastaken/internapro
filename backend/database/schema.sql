-- INTERNA - Internship Management System Database Schema
-- MySQL Database Design

CREATE DATABASE IF NOT EXISTS interna_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE interna_db;

-- Users table (for authentication)
CREATE TABLE users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified_at TIMESTAMP NULL,
    password_hash VARCHAR(255) NULL,
    role ENUM('admin', 'supervisor', 'pembimbing', 'intern') NOT NULL,
    avatar VARCHAR(255),

    -- OAuth / Provider
    google_id VARCHAR(255) UNIQUE,
    provider VARCHAR(50),

    -- 2FA (legacy TOTP + Fortify style)
    totp_secret VARCHAR(255), -- for Google Authenticator
    is_2fa_enabled BOOLEAN DEFAULT FALSE,
    google2fa_secret VARCHAR(255),
    google2fa_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret TEXT,
    two_factor_recovery_codes TEXT,
    two_factor_confirmed_at TIMESTAMP NULL,

    remember_token VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_role (role)
) ENGINE=InnoDB;

-- Institutions table
CREATE TABLE institutions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    phone VARCHAR(50),
    email VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- Supervisors table
CREATE TABLE supervisors (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    nip VARCHAR(50),
    phone VARCHAR(50),
    position VARCHAR(100),
    address TEXT,
    institution VARCHAR(255),
    status ENUM('pending', 'active') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_status (status)
) ENGINE=InnoDB;

-- Interns table (siswa magang)
CREATE TABLE interns (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    institution_id BIGINT,
    supervisor_id BIGINT,
    full_name VARCHAR(255) NOT NULL,
    nis VARCHAR(50),
    student_id VARCHAR(50),
    school VARCHAR(255),
    department VARCHAR(255),
    date_of_birth DATE,
    gender ENUM('male', 'female'),
    phone VARCHAR(50),
    address TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status ENUM('pending', 'active', 'completed', 'cancelled', 'terminated') DEFAULT 'active',
    certificate_number VARCHAR(100),
    certificate_issued_at DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (institution_id) REFERENCES institutions(id) ON DELETE RESTRICT,
    FOREIGN KEY (supervisor_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_status (status),
    INDEX idx_dates (start_date, end_date),
    INDEX idx_supervisor (supervisor_id)
) ENGINE=InnoDB;

-- Task assignments (bulk assignment)
CREATE TABLE task_assignments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    intern_id BIGINT NULL, -- optional single-intern assignment (used by dashboard queries)
    title VARCHAR(255) NOT NULL,
    description TEXT,
    assigned_by BIGINT NOT NULL, -- user_id (admin/pembimbing)
    priority ENUM('low', 'medium', 'high') DEFAULT 'medium',
    status ENUM('pending', 'scheduled', 'in_progress', 'submitted', 'revision', 'completed', 'overdue') DEFAULT 'pending',
    start_date DATE,
    deadline DATE,
    deadline_time TIME,
    submitted_at DATETIME,
    is_late BOOLEAN DEFAULT FALSE,
    grade INT,
    assign_to_all BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (intern_id) REFERENCES interns(id) ON DELETE CASCADE,
    INDEX idx_assigned_by (assigned_by),
    INDEX idx_priority (priority),
    INDEX idx_deadline (deadline),
    INDEX idx_intern_status (intern_id, status)
) ENGINE=InnoDB;

CREATE TABLE task_assignment_interns (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_assignment_id BIGINT NOT NULL,
    intern_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (task_assignment_id) REFERENCES task_assignments(id) ON DELETE CASCADE,
    FOREIGN KEY (intern_id) REFERENCES interns(id) ON DELETE CASCADE,
    UNIQUE KEY uniq_assignment_intern (task_assignment_id, intern_id)
) ENGINE=InnoDB;

-- Tasks table (penugasan)
CREATE TABLE tasks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_assignment_id BIGINT,
    intern_id BIGINT,
    assigned_by BIGINT NOT NULL, -- user_id (admin/pembimbing)
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority ENUM('low', 'medium', 'high') DEFAULT 'medium',
    status ENUM('pending', 'scheduled', 'in_progress', 'submitted', 'revision', 'completed', 'overdue') DEFAULT 'pending',
    start_date DATE,
    deadline DATE,
    deadline_time TIME,
    started_at DATETIME,
    submitted_at DATETIME,
    completed_at DATETIME,
    approved_at TIMESTAMP NULL,
    is_late BOOLEAN DEFAULT FALSE,
    submission_notes TEXT,
    submission_links JSON,
    score INT,
    admin_feedback TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (task_assignment_id) REFERENCES task_assignments(id) ON DELETE SET NULL,
    FOREIGN KEY (intern_id) REFERENCES interns(id) ON DELETE SET NULL,
    FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_assigned_by (assigned_by),
    INDEX idx_intern_status (intern_id, status),
    INDEX idx_deadline (deadline)
) ENGINE=InnoDB;

-- Task attachments (upload foto/pdf)
CREATE TABLE task_attachments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_type ENUM('jpg', 'jpeg', 'png', 'pdf') NOT NULL,
    file_size BIGINT NOT NULL, -- in bytes
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    INDEX idx_task (task_id)
) ENGINE=InnoDB;

-- Office settings (for attendance geolocation)
CREATE TABLE office_settings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    radius_meters INT DEFAULT 1000, -- 1km default
    check_in_time TIME NOT NULL, -- jam hadir
    check_out_time TIME NOT NULL, -- jam pulang
    late_tolerance_minutes INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- Attendance table (presensi)
CREATE TABLE attendances (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    intern_id BIGINT NOT NULL,
    date DATE NOT NULL,
    check_in_time DATETIME,
    check_in_latitude DECIMAL(10, 8),
    check_in_longitude DECIMAL(11, 8),
    check_out_time DATETIME,
    check_out_latitude DECIMAL(10, 8),
    check_out_longitude DECIMAL(11, 8),
    status ENUM('present', 'late', 'absent', 'sick', 'permission', 'on_leave', 'excused') DEFAULT 'absent',
    late_reason TEXT,
    notes TEXT,
    distance_meters INT,
    proof_file VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (intern_id) REFERENCES interns(id) ON DELETE CASCADE,
    UNIQUE KEY unique_intern_date (intern_id, date),
    INDEX idx_date (date),
    INDEX idx_status (status),
    INDEX idx_status_date (status, date)
) ENGINE=InnoDB;

-- Leave permissions (perizinan)
CREATE TABLE leave_requests (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    intern_id BIGINT NOT NULL,
    leave_type ENUM('sick', 'permission', 'other') NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT NOT NULL,
    attachment_path VARCHAR(500), -- upload foto surat izin
    status ENUM('pending', 'approved', 'rejected') DEFAULT 'pending',
    approved_by BIGINT, -- user_id (admin/pembimbing)
    approved_at DATETIME,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (intern_id) REFERENCES interns(id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_intern_status (intern_id, status),
    INDEX idx_dates (start_date, end_date)
) ENGINE=InnoDB;

-- Assessments (penilaian)
CREATE TABLE assessments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    intern_id BIGINT NOT NULL,
    task_id BIGINT, -- optional, can be task-specific or general
    assessed_by BIGINT NOT NULL, -- user_id (admin/pembimbing)
    score INT NOT NULL CHECK (score >= 0 AND score <= 100),
    category ENUM('very_good', 'good', 'not_good', 'very_bad') GENERATED ALWAYS AS (
        CASE
            WHEN score >= 85 THEN 'very_good'
            WHEN score >= 70 THEN 'good'
            WHEN score >= 50 THEN 'not_good'
            ELSE 'very_bad'
        END
    ) STORED,
    aspect VARCHAR(100) NOT NULL, -- e.g., 'discipline', 'work_quality', 'attitude'
    -- Laravel-style multi-criteria scores
    quality_score INT,
    speed_score INT,
    initiative_score INT,
    teamwork_score INT,
    communication_score INT,
    strengths TEXT,
    improvements TEXT,
    comments TEXT,
    notes TEXT,
    assessment_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (intern_id) REFERENCES interns(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE SET NULL,
    FOREIGN KEY (assessed_by) REFERENCES users(id) ON DELETE RESTRICT,
    INDEX idx_intern (intern_id),
    INDEX idx_category (category)
) ENGINE=InnoDB;

-- Reports (weekly/monthly/final)
CREATE TABLE reports (
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
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_intern_type (intern_id, type)
) ENGINE=InnoDB;

-- Settings (key/value)
CREATE TABLE settings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    `key` VARCHAR(255) UNIQUE NOT NULL,
    `value` TEXT,
    `type` VARCHAR(50) DEFAULT 'string',
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- Notifications (in-app)
CREATE TABLE notifications (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    icon VARCHAR(255),
    link VARCHAR(255),
    data JSON,
    read_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_read (user_id, read_at),
    INDEX idx_user_created (user_id, created_at),
    INDEX idx_type (type)
) ENGINE=InnoDB;

-- Reports/Certificates
CREATE TABLE certificates (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    intern_id BIGINT NOT NULL,
    certificate_number VARCHAR(100) UNIQUE NOT NULL,
    issue_date DATE NOT NULL,
    final_score DECIMAL(5, 2),
    remarks TEXT,
    file_path VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (intern_id) REFERENCES interns(id) ON DELETE CASCADE,
    INDEX idx_intern (intern_id)
) ENGINE=InnoDB;

-- Activity logs (for audit trail)
CREATE TABLE activity_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50), -- e.g., 'task', 'attendance', 'assessment'
    entity_id BIGINT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_user_created (user_id, created_at),
    INDEX idx_entity (entity_type, entity_id)
) ENGINE=InnoDB;

-- Password reset tokens
CREATE TABLE password_resets (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_created (created_at)
) ENGINE=InnoDB;

-- Insert default office settings
-- PT. Duta Solusi Informatika, Semarang
INSERT INTO office_settings (latitude, longitude, radius_meters, check_in_time, check_out_time, late_tolerance_minutes)
VALUES (-7.035549620262833, 110.47464898482643, 1000, '08:30:00', '16:00:00', 15);
