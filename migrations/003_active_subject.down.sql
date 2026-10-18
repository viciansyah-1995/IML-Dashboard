DROP INDEX report_active_subject;
CREATE UNIQUE INDEX report_active_subject ON report_versions(subject_key,scope,basis,currency,definition_version) WHERE active;
