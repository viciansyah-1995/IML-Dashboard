ALTER TABLE users ADD COLUMN joined_on date;
UPDATE users SET joined_on=(created_at AT TIME ZONE 'Asia/Jakarta')::date;
ALTER TABLE users ALTER COLUMN joined_on SET NOT NULL;
ALTER TABLE users ADD COLUMN photo_data bytea, ADD COLUMN photo_type text NOT NULL DEFAULT '', ADD COLUMN photo_version text NOT NULL DEFAULT '';
ALTER TABLE hosts ADD COLUMN full_name text NOT NULL DEFAULT '', ADD COLUMN email text NOT NULL DEFAULT '', ADD COLUMN birth_date date, ADD COLUMN gender text NOT NULL DEFAULT '' CHECK(gender IN ('','female','male','other','undisclosed')), ADD COLUMN joined_on date, ADD COLUMN city text NOT NULL DEFAULT '', ADD COLUMN languages text NOT NULL DEFAULT '', ADD COLUMN specialties text NOT NULL DEFAULT '', ADD COLUMN photo_data bytea, ADD COLUMN photo_type text NOT NULL DEFAULT '', ADD COLUMN photo_version text NOT NULL DEFAULT '';
UPDATE hosts SET full_name=name;
CREATE INDEX hosts_profile_reporting ON hosts(city,gender,joined_on);
