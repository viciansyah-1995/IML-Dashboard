DROP INDEX hosts_profile_reporting;
ALTER TABLE hosts DROP COLUMN full_name, DROP COLUMN email, DROP COLUMN birth_date, DROP COLUMN gender, DROP COLUMN joined_on, DROP COLUMN city, DROP COLUMN languages, DROP COLUMN specialties, DROP COLUMN photo_data, DROP COLUMN photo_type, DROP COLUMN photo_version;
ALTER TABLE users DROP COLUMN joined_on, DROP COLUMN photo_data, DROP COLUMN photo_type, DROP COLUMN photo_version;
