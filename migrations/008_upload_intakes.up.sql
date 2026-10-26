CREATE TABLE upload_intakes (
 id text PRIMARY KEY, created_by text NOT NULL REFERENCES users(id),
 channel text NOT NULL CHECK(channel IN('tiktok','shopee')),
 mapping_version text NOT NULL REFERENCES metric_mapping_versions(id),
 status text NOT NULL DEFAULT 'uploaded' CHECK(status IN('uploaded','queued','running','extracted','failed','submitted')),
 step integer NOT NULL DEFAULT 1 CHECK(step BETWEEN 1 AND 4),
 revision integer NOT NULL DEFAULT 1, metadata jsonb NOT NULL DEFAULT '{}',
 result jsonb, corrections jsonb NOT NULL DEFAULT '[]',
 attempts integer NOT NULL DEFAULT 0, lease_token text, lease_until timestamptz,
 next_run_at timestamptz NOT NULL DEFAULT now(), last_error text NOT NULL DEFAULT '',
 idempotency_key text NOT NULL UNIQUE, report_id text REFERENCES report_versions(id),
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX upload_intakes_owner ON upload_intakes(created_by,created_at DESC);
CREATE INDEX upload_intakes_queue ON upload_intakes(next_run_at) WHERE status IN('queued','running');
CREATE TABLE intake_attachments (
 id text PRIMARY KEY, intake_id text NOT NULL REFERENCES upload_intakes(id),
 original_name text NOT NULL, storage_key text NOT NULL UNIQUE,
 content_type text NOT NULL CHECK(content_type IN('image/png','image/jpeg')),
 content_hash text NOT NULL, bytes bigint NOT NULL CHECK(bytes>0)
);
-- Host accounts can only upload their own drafts and are explicitly linked by Superadmin.
INSERT INTO roles(name) VALUES('host');
INSERT INTO role_permissions(role,permission) VALUES('host','upload_report');
ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK(role IN('superadmin','admin','client','host'));
ALTER TABLE users ADD COLUMN host_id text REFERENCES hosts(id);
ALTER TABLE users ADD CONSTRAINT users_host_link CHECK(role<>'host' OR host_id IS NOT NULL);
