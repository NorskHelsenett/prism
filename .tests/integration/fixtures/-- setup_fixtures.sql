-- setup_fixtures.sql
-- Create tables and insert data for prism.db, updated with minimum UI fields

-- Create user_data table
CREATE TABLE IF NOT EXISTS user_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    email TEXT,
    name TEXT,
    picture TEXT,
    role TEXT DEFAULT 'visitor',
    title TEXT DEFAULT 'My title',
    otp_secret TEXT,
    notifications JSON,
    settings JSON,
    UNIQUE(email)
);

-- Create project_data table
CREATE TABLE IF NOT EXISTS project_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    project_name TEXT NOT NULL,
    slack_channel TEXT,
    description TEXT,
    client_email TEXT,
    hacker_name TEXT,
    is_bug_bounty NUMERIC
);

-- Create json_data table
CREATE TABLE IF NOT EXISTS json_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    vulnerability JSON,
    found_by TEXT,
    project_id INTEGER,
    status TEXT DEFAULT 'Reported',
    slack_url TEXT,
    comments JSON,
    revisions JSON,
    FOREIGN KEY (project_id) REFERENCES project_data(id)
);

-- Create event_queues table (required for triggers)
CREATE TABLE IF NOT EXISTS event_queues (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    table_id INTEGER,
    table_name TEXT,
    error TEXT,
    processed NUMERIC DEFAULT false,
    created_at DATETIME,
    updated_at DATETIME,
    kind INTEGER
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_user_data_deleted_at ON user_data(deleted_at);
CREATE INDEX IF NOT EXISTS idx_project_data_deleted_at ON project_data(deleted_at);
CREATE INDEX IF NOT EXISTS idx_json_data_deleted_at ON json_data(deleted_at);
CREATE INDEX IF NOT EXISTS idx_created_at ON event_queues(created_at);
CREATE INDEX IF NOT EXISTS idx_processed ON event_queues(processed);

-- Create triggers for json_data
CREATE TRIGGER IF NOT EXISTS jsondata_insert AFTER INSERT ON json_data
BEGIN
    INSERT INTO event_queues (table_id, table_name, created_at, kind)
    VALUES (NEW.id, 'vulnerability', CURRENT_TIMESTAMP, 1);
END;

CREATE TRIGGER IF NOT EXISTS jsondata_update_comments
AFTER UPDATE OF comments ON json_data
FOR EACH ROW
WHEN (OLD.comments IS NOT NEW.comments)
BEGIN
    INSERT INTO event_queues (table_id, table_name, created_at, kind)
    VALUES (NEW.id, 'vulnerability', CURRENT_TIMESTAMP, 2);
END;

-- Insert user_data
INSERT INTO user_data (email, role, name, created_at, updated_at) VALUES
('alice.admin@test.local', 'admin', 'Alice Admin', '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
('bob.viewer@test.local', 'pentester', 'Bob Pentester', '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
('charlie.visitor@test.local', 'visitor', 'Charlie Visitor', '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
('diana.external@test.local', 'PET01', 'Diana External Pentester', '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
('eve.visitor@test.local', 'visitor', 'Eve Visitor', '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
('frank.viewer@test.local', 'global_viewer', 'Frank Viewer', '2025-10-17 09:00:00', '2025-10-17 09:00:00');

-- Insert project_data
INSERT INTO project_data (id, project_name, description, slack_channel, client_email, hacker_name, is_bug_bounty, created_at, updated_at) VALUES
(1, 'ProjectAlpha', 'Test project for access control testing', '#project-alpha', 'charlie.visitor@test.local', 'bob.viewer@test.local', 0, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(2, 'ProjectBeta', 'Another test project', '#project-beta', '', 'diana.external@test.local', 1, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(3, 'ProjectGamma', 'Public project for testing', '#project-gamma', 'eve.visitor@test.local', '', 0, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(4, 'SIGMA', 'Sigma project for access control testing', '#sigma-channel', ',bob.viewer@test.local,eve.visitor@test.local', ',diana.external@test.local', 1, '2025-10-17 09:00:00', '2025-10-17 09:00:00');

-- Insert json_data (vulnerabilities)
INSERT INTO json_data (id, vulnerability, found_by, project_id, status, slack_url, comments, revisions, created_at, updated_at) VALUES
(1, '{"title": "SQL Injection in Login Form", "criticality": "critical", "evidence": "Allows authentication bypass via SQL injection", "date": "2025-10-16", "visibility": "hidden", "assignedTo": null}', 'bob.viewer@test.local', 1, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(2, '{"title": "XSS in User Profile", "criticality": "medium", "evidence": "Reflected XSS in user profile page", "date": "2025-10-16", "visibility": "undisclosed", "assignedTo": "charlie.visitor@test.local"}', 'bob.viewer@test.local', 1, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(3, '{"title": "CSRF Token Missing", "criticality": "low", "evidence": "Missing CSRF protection on password change", "date": "2025-10-16", "visibility": "hidden", "assignedTo": null}', 'diana.external@test.local', 2, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(4, '{"title": "Authentication Bypass via JWT", "criticality": "critical", "evidence": "JWT signature validation flaw", "date": "2025-10-16", "visibility": "published", "assignedTo": null}', 'diana.external@test.local', 2, 'Fixed', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(5, '{"title": "Information Disclosure in API", "criticality": "low", "evidence": "API leaks sensitive user info", "date": "2025-10-16", "visibility": "published", "assignedTo": null}', '', 3, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(6, '{"title": "Private Finding - Under Investigation", "criticality": "high", "evidence": "Sensitive finding under analysis", "date": "2025-10-16", "visibility": "published", "assignedTo": "bob.viewer@test.local"}', 'bob.viewer@test.local', 1, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(7, '{"title": "Race Condition in Payment", "criticality": "high", "evidence": "Race condition allows double-spend", "date": "2025-10-16", "visibility": "undisclosed", "assignedTo": "diana.external@test.local"}', 'bob.viewer@test.local', 1, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(8, '{"title": "Insecure Direct Object Reference", "criticality": "medium", "evidence": "IDOR in document download endpoint", "date": "2025-10-16", "visibility": "published", "assignedTo": null}', 'diana.external@test.local', 2, 'Open', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(9, '{"title": "Privilege Escalation via Misconfigured Roles", "criticality": "high", "evidence": "Misconfigured roles allow privilege escalation", "date": "2025-10-16", "visibility": "published", "assignedTo": null}', 'bob.viewer@test.local', NULL, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(10, '{"title": "Unvalidated Redirects", "criticality": "medium", "evidence": "Redirects allow phishing attacks", "date": "2025-10-16", "visibility": "undisclosed", "assignedTo": null}', 'bob.viewer@test.local', 2, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(11, '{"title": "Sensitive Data Exposure in Logs", "criticality": "low", "evidence": "Logs contain plaintext sensitive data", "date": "2025-10-16", "visibility": "published", "assignedTo": null}', 'diana.external@test.local', NULL, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(12, '{"title": "SQL Injection in SIGMA Login", "criticality": "critical", "evidence": "Bypassed authentication with '' or ''='' via login form", "date": "2025-10-16", "visibility": "undisclosed", "assignedTo": "bob.viewer@test.local"}', 'diana.external@test.local', 4, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(13, '{"title": "XSS in SIGMA Dashboard", "criticality": "medium", "evidence": "<script>alert(''xss'')</script> executed in dashboard input", "date": "2025-10-16", "visibility": "published", "assignedTo": null}', 'bob.viewer@test.local', 4, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00'),
(14, '{"title": "SIGMA API Key Leak", "criticality": "information", "evidence": "API keys exposed in response headers", "date": "2025-10-16", "visibility": "undisclosed", "assignedTo": "eve.visitor@test.local"}', 'diana.external@test.local', 4, 'Reported', NULL, NULL, NULL, '2025-10-17 09:00:00', '2025-10-17 09:00:00');

-- Recreate the accessible_vulnerabilities view
DROP VIEW IF EXISTS accessible_vulnerabilities;
CREATE VIEW accessible_vulnerabilities AS
SELECT
    json_data.id AS id,
    json_extract(json_data.vulnerability, '$.visibility') AS visibility,
    json_extract(json_data.vulnerability, '$.assignedTo') AS assigned_to,
    json_data.found_by,
    json_data.project_id,
    project_data.client_email,
    project_data.hacker_name
FROM json_data
LEFT JOIN project_data ON json_data.project_id = project_data.id;