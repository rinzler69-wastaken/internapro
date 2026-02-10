-- Auto-generated SQL Schema
-- Generated from Go backend models

CREATE TABLE assessment (
  i_d BIGINT NOT NULL,
  intern_i_d BIGINT NOT NULL,
  task_i_d VARCHAR(255) NOT NULL,
  assessed_by BIGINT NOT NULL,
  score INTEGER NOT NULL,
  category VARCHAR(255) NOT NULL,
  aspect VARCHAR(255) NOT NULL,
  quality_score VARCHAR(255) NOT NULL,
  speed_score VARCHAR(255) NOT NULL,
  initiative_score VARCHAR(255) NOT NULL,
  teamwork_score VARCHAR(255) NOT NULL,
  communication_score VARCHAR(255) NOT NULL,
  strengths VARCHAR(255) NOT NULL,
  improvements VARCHAR(255) NOT NULL,
  comments VARCHAR(255) NOT NULL,
  notes VARCHAR(255) NOT NULL,
  assessment_date TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  intern_name VARCHAR(255) NOT NULL,
  assessor_name VARCHAR(255) NOT NULL,
  task_title VARCHAR(255) NOT NULL
);

CREATE TABLE attendance (
  i_d BIGINT NOT NULL,
  intern_i_d BIGINT NOT NULL,
  date TIMESTAMP NOT NULL,
  check_in_time TIMESTAMP NOT NULL,
  check_in_latitude DOUBLE NOT NULL,
  check_in_longitude DOUBLE NOT NULL,
  check_out_time TIMESTAMP NOT NULL,
  check_out_latitude DOUBLE NOT NULL,
  check_out_longitude DOUBLE NOT NULL,
  status VARCHAR(255) NOT NULL,
  late_reason VARCHAR(255) NOT NULL,
  notes VARCHAR(255) NOT NULL,
  distance_meters INTEGER NOT NULL,
  proof_file VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  intern_name VARCHAR(255) NOT NULL
);

CREATE TABLE office_settings (
  i_d BIGINT NOT NULL,
  latitude DOUBLE NOT NULL,
  longitude DOUBLE NOT NULL,
  radius_meters INTEGER NOT NULL,
  check_in_time VARCHAR(255) NOT NULL,
  check_out_time VARCHAR(255) NOT NULL,
  late_tolerance_minutes INTEGER NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE TABLE intern (
  i_d BIGINT NOT NULL,
  user_i_d BIGINT NOT NULL,
  institution_i_d VARCHAR(255) NOT NULL,
  supervisor_i_d VARCHAR(255) NOT NULL,
  full_name VARCHAR(255) NOT NULL,
  n_i_s VARCHAR(255) NOT NULL,
  student_i_d VARCHAR(255) NOT NULL,
  school VARCHAR(255) NOT NULL,
  department VARCHAR(255) NOT NULL,
  date_of_birth VARCHAR(255) NOT NULL,
  gender VARCHAR(255) NOT NULL,
  phone VARCHAR(255) NOT NULL,
  address VARCHAR(255) NOT NULL,
  start_date TIMESTAMP NOT NULL,
  end_date TIMESTAMP NOT NULL,
  status VARCHAR(255) NOT NULL,
  certificate_number VARCHAR(255) NOT NULL,
  certificate_issued_at VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  user VARCHAR(255) NOT NULL,
  institution VARCHAR(255) NOT NULL,
  supervisor VARCHAR(255) NOT NULL
);

CREATE TABLE leave_request (
  i_d BIGINT NOT NULL,
  intern_i_d BIGINT NOT NULL,
  leave_type VARCHAR(255) NOT NULL,
  start_date TIMESTAMP NOT NULL,
  end_date TIMESTAMP NOT NULL,
  reason VARCHAR(255) NOT NULL,
  attachment_path VARCHAR(255) NOT NULL,
  status VARCHAR(255) NOT NULL,
  approved_by BIGINT NOT NULL,
  approved_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  intern_name VARCHAR(255) NOT NULL,
  approver_name VARCHAR(255) NOT NULL
);

CREATE TABLE notification (
  i_d BIGINT NOT NULL,
  user_i_d BIGINT NOT NULL,
  type VARCHAR(255) NOT NULL,
  title VARCHAR(255) NOT NULL,
  message VARCHAR(255) NOT NULL,
  icon VARCHAR(255) NOT NULL,
  link VARCHAR(255) NOT NULL,
  data VARCHAR(255) NOT NULL,
  read_at VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE TABLE report (
  i_d BIGINT NOT NULL,
  intern_i_d BIGINT NOT NULL,
  created_by BIGINT NOT NULL,
  title VARCHAR(255) NOT NULL,
  content VARCHAR(255) NOT NULL,
  type VARCHAR(255) NOT NULL,
  period_start TIMESTAMP NOT NULL,
  period_end TIMESTAMP NOT NULL,
  status VARCHAR(255) NOT NULL,
  feedback VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  intern_name VARCHAR(255) NOT NULL,
  created_by_name VARCHAR(255) NOT NULL
);

CREATE TABLE setting (
  i_d BIGINT NOT NULL,
  key VARCHAR(255) NOT NULL,
  value VARCHAR(255) NOT NULL,
  type VARCHAR(255) NOT NULL,
  description VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE TABLE task (
  i_d BIGINT NOT NULL,
  task_assignment_i_d BIGINT NOT NULL,
  intern_i_d BIGINT NOT NULL,
  assigned_by BIGINT NOT NULL,
  title VARCHAR(255) NOT NULL,
  description VARCHAR(255) NOT NULL,
  priority VARCHAR(255) NOT NULL,
  status VARCHAR(255) NOT NULL,
  start_date TIMESTAMP NOT NULL,
  deadline TIMESTAMP NOT NULL,
  deadline_time VARCHAR(255) NOT NULL,
  started_at TIMESTAMP NOT NULL,
  submitted_at TIMESTAMP NOT NULL,
  completed_at TIMESTAMP NOT NULL,
  approved_at TIMESTAMP NOT NULL,
  is_late BOOLEAN NOT NULL,
  submission_notes VARCHAR(255) NOT NULL,
  submission_links JSON NOT NULL,
  score INTEGER NOT NULL,
  admin_feedback VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  intern_name VARCHAR(255) NOT NULL,
  assigned_by_name VARCHAR(255) NOT NULL
);

CREATE TABLE task_assignment (
  i_d BIGINT NOT NULL,
  title VARCHAR(255) NOT NULL,
  description VARCHAR(255) NOT NULL,
  assigned_by BIGINT NOT NULL,
  priority VARCHAR(255) NOT NULL,
  start_date TIMESTAMP NOT NULL,
  deadline TIMESTAMP NOT NULL,
  deadline_time VARCHAR(255) NOT NULL,
  assign_to_all BOOLEAN NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  assigned_by_name VARCHAR(255) NOT NULL,
  tasks_count INTEGER NOT NULL
);

CREATE TABLE task_attachment (
  i_d BIGINT NOT NULL,
  task_i_d BIGINT NOT NULL,
  file_name VARCHAR(255) NOT NULL,
  file_path VARCHAR(255) NOT NULL,
  file_type VARCHAR(255) NOT NULL,
  file_size BIGINT NOT NULL,
  uploaded_at TIMESTAMP NOT NULL
);

CREATE TABLE user (
  i_d BIGINT NOT NULL,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  email_verified_at VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(255) NOT NULL,
  avatar VARCHAR(255) NOT NULL,
  google_i_d VARCHAR(255) NOT NULL,
  provider VARCHAR(255) NOT NULL,
  t_o_t_p_secret VARCHAR(255) NOT NULL,
  is2_f_a_enabled BOOLEAN NOT NULL,
  google2_f_a_secret VARCHAR(255) NOT NULL,
  google2_f_a_enabled BOOLEAN NOT NULL,
  two_factor_secret VARCHAR(255) NOT NULL,
  two_factor_recovery VARCHAR(255) NOT NULL,
  two_factor_confirmed VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE TABLE supervisor (
  i_d BIGINT NOT NULL,
  user_i_d BIGINT NOT NULL,
  full_name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  n_i_p VARCHAR(255) NOT NULL,
  phone VARCHAR(255) NOT NULL,
  position VARCHAR(255) NOT NULL,
  address VARCHAR(255) NOT NULL,
  institution VARCHAR(255) NOT NULL,
  status VARCHAR(255) NOT NULL,
  interns_count BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE TABLE institution (
  i_d BIGINT NOT NULL,
  name VARCHAR(255) NOT NULL,
  address VARCHAR(255) NOT NULL,
  phone VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

