-- Migrate assessed_by FK to reference users(id) instead of supervisors(id)
ALTER TABLE assessments
  DROP FOREIGN KEY assessments_ibfk_3;

ALTER TABLE assessments
  ADD CONSTRAINT assessments_assessed_by_fk
  FOREIGN KEY (assessed_by) REFERENCES users(id) ON DELETE RESTRICT;
