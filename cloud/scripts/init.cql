-- Create keyspace
CREATE KEYSPACE IF NOT EXISTS cloud_storage
    WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE cloud_storage;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id text PRIMARY KEY,
    email text,
    name text,
    avatar text,
    created_at timestamp,
    updated_at timestamp
);
CREATE INDEX IF NOT EXISTS users_email_idx ON users(email);

-- Files table
CREATE TABLE IF NOT EXISTS files (
    id text PRIMARY KEY,
    name text,
    size bigint,
    mime_type text,
    url text,
    user_id text,
    created_at timestamp,
    updated_at timestamp
);
CREATE INDEX IF NOT EXISTS files_user_id_idx ON files(user_id);

-- Notes table
CREATE TABLE IF NOT EXISTS notes (
    id text PRIMARY KEY,
    title text,
    content text,
    user_id text,
    created_at timestamp,
    updated_at timestamp
);
CREATE INDEX IF NOT EXISTS notes_user_id_idx ON notes(user_id);
