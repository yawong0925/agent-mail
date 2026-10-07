// database/schema.go
package database

// schemaEmails defines the structure for DB1 (Emails & Attachments).
const schemaEmails = `
CREATE TABLE IF NOT EXISTS emails (
	id INTEGER PRIMARY KEY AUTOINCREMENT, 
	account_id INTEGER, 
	remote_uid TEXT,
	date_time DATETIME, 
	sender TEXT, 
	recipient TEXT, 
	subject TEXT,
	has_attachment BOOLEAN DEFAULT 0, 
	eml_file_path TEXT, 
	is_processed BOOLEAN DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_search ON emails(date_time, sender, subject, recipient);

CREATE TABLE IF NOT EXISTS email_attachments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email_id INTEGER,
	original_filename TEXT,
	stored_file_path TEXT,
	mime_type TEXT,
	file_size INTEGER,
	FOREIGN KEY(email_id) REFERENCES emails(id)
);
`

// schemaAuth defines the structure for DB2 (API Tokens & Access Logs).
// Notice that 'portal_sessions' is intentionally excluded because we use RAM for that now.
const schemaAuth = `
CREATE TABLE IF NOT EXISTS api_tokens (
	id INTEGER PRIMARY KEY AUTOINCREMENT, 
	user_id INTEGER,
	agent_name TEXT,
	token TEXT UNIQUE,
	rate_limit_per_min INTEGER DEFAULT 10, 
	start_date DATETIME,
	expires_at DATETIME, 
	is_revoked BOOLEAN DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS access_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT, 
	ip_address TEXT, 
	token_id INTEGER,
	agent_name TEXT,
	endpoint TEXT, 
	email_id_accessed INTEGER, 
	status TEXT, 
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS failed_logins (
	ip_address TEXT PRIMARY KEY, 
	attempts INTEGER DEFAULT 1, 
	blocked_until DATETIME
);
`

// schemaMgmt defines the structure for DB3 (Users, Telemetry, & Settings).
// This includes the heavily expanded imap_accounts table.
const schemaMgmt = `
CREATE TABLE IF NOT EXISTS local_users (
	id INTEGER PRIMARY KEY AUTOINCREMENT, 
	username TEXT UNIQUE, 
	email TEXT UNIQUE,
	password_hash TEXT,
	two_factor_code TEXT,       -- 6-char alphanumeric code
	two_factor_expires DATETIME -- Expiration (5 mins)
);
CREATE TABLE IF NOT EXISTS imap_accounts (
	id INTEGER PRIMARY KEY AUTOINCREMENT, 
	user_id INTEGER, 
	email_address TEXT,
	imap_host TEXT, 
	imap_port INTEGER, 
	encryption TEXT DEFAULT 'tls', 
	username TEXT, 
	password_encrypted TEXT,
	is_active BOOLEAN DEFAULT 1,
	push_supported BOOLEAN DEFAULT 0,  -- 1 if IMAP IDLE is supported
	polling_interval INTEGER DEFAULT 5, -- Minimum 5 minutes
	bandwidth_in_bytes INTEGER DEFAULT 0,
	bandwidth_out_bytes INTEGER DEFAULT 0,
	total_bandwidth_in_bytes INTEGER DEFAULT 0,
	total_bandwidth_out_bytes INTEGER DEFAULT 0,
	last_email_datetime DATETIME,
	last_check_datetime DATETIME,
	last_error TEXT,
	last_agent_name TEXT,
	last_agent_access DATETIME
);
CREATE TABLE IF NOT EXISTS oauth_accounts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER,
	email_address TEXT,
	provider TEXT,
	access_token TEXT,
	refresh_token TEXT,
	expires_at DATETIME,
	is_active BOOLEAN DEFAULT 1
);
CREATE TABLE IF NOT EXISTS ip_blacklist (
	ip_address TEXT PRIMARY KEY, 
	reason TEXT, 
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`