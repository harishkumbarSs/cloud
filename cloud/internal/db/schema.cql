-- Create keyspace
CREATE KEYSPACE IF NOT EXISTS cloud_storage
    WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE cloud_storage;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    email text PRIMARY KEY,
    name text,
    created_at timestamp,
    last_login timestamp
);

-- Files metadata table
CREATE TABLE IF NOT EXISTS files (
    user_email text,
    file_id uuid,
    filename text,
    size bigint,
    content_type text,
    storage_path text,
    uploaded_at timestamp,
    PRIMARY KEY ((user_email), file_id)
);

-- Notes table
CREATE TABLE IF NOT EXISTS notes (
    user_email text,
    note_id uuid,
    title text,
    content text,
    created_at timestamp,
    updated_at timestamp,
    PRIMARY KEY ((user_email), note_id)
);
