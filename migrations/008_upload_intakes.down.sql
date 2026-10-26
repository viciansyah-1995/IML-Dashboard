DROP TABLE intake_attachments;
DROP TABLE upload_intakes;
ALTER TABLE users DROP CONSTRAINT users_host_link;
ALTER TABLE users DROP COLUMN host_id;
DELETE FROM role_permissions WHERE role='host';
DELETE FROM roles WHERE name='host';
ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK(role IN('superadmin','admin','client'));
