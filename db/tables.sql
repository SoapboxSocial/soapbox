CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    display_name VARCHAR(256) NOT NULL,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(254) NOT NULL,
    image VARCHAR(100) NOT NULL DEFAULT '',
    bio TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX idx_email ON users (email);
CREATE UNIQUE INDEX idx_username ON users (username);

CREATE TABLE IF NOT EXISTS apple_authentication (
    user_id INT NOT NULL,
    apple_user TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_apple_authentication ON apple_authentication (user_id);

CREATE TABLE IF NOT EXISTS followers (
    follower INT NOT NULL,
    user_id INT NOT NULL,
    CHECK (user_id != follower),
    FOREIGN KEY (follower) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (follower, user_id)
);

-- Check token length
CREATE TABLE IF NOT EXISTS devices (
    token VARCHAR(64) PRIMARY KEY,
    user_id INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS linked_accounts (
    user_id INT NOT NULL,
    provider VARCHAR(7) NOT NULL,
    profile_id BIGINT NOT NULl,
    token TEXT NOT NULL DEFAULT '',
    secret TEXT NOT NULL DEFAULT '',
    username VARCHAR(256) NOT NULL DEFAULT '',
    CHECK (provider IN ('twitter')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_profiles ON linked_accounts (provider, profile_id);
CREATE UNIQUE INDEX idx_provider ON linked_accounts (provider, user_id);

CREATE TABLE IF NOT EXISTS group_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(30)
);

INSERT INTO group_types (name) VALUES ('public'), ('private'), ('restricted');

CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    image VARCHAR(100) NOT NULL DEFAULT '',
    group_type INT NOT NULL,
    FOREIGN KEY (group_type) REFERENCES group_types(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_group_name ON groups (lower(name));

CREATE TABLE IF NOT EXISTS group_members (
    group_id INT NOT NULL,
    user_id INT NOT NULL,
    role VARCHAR(5) NOT NULL,
    CHECK (role IN ('admin', 'user')),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_group_membership ON group_members (group_id, user_id);

CREATE TABLE IF NOT EXISTS group_invites (
    group_id INT NOT NULL,
    from_id INT NOT NULL,
    user_id INT NOT NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (from_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_group_invites ON group_invites (group_id, user_id);

CREATE TABLE IF NOT EXISTS stories (
    id VARCHAR(256) PRIMARY KEY,
    user_id INT NOT NULL,
    expires_at INT NOT NULL,
    device_timestamp INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_stories ON stories (id, user_id);

CREATE TABLE IF NOT EXISTS story_reactions (
    story_id VARCHAR(256) NOT NULL,
    user_id INT NOT NULL,
    reaction VARCHAR(10) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (story_id) REFERENCES stories(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_stories_react ON story_reactions (story_id, user_id);

CREATE TABLE IF NOT EXISTS blocks (
    user_id INT NOT NULL,
    blocked INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (blocked) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_blocks ON blocks (user_id, blocked);

CREATE TABLE IF NOT exists mini_developers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE UNIQUE INDEX idx_mini_developers_name ON mini_developers (name);

CREATE TABLE IF NOT EXISTS minis (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    image VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    developer_id INT NOT NULL,
    FOREIGN KEY (developer_id) REFERENCES mini_developers(id) ON DELETE CASCADE
);

-- Inserting apps
INSERT INTO mini_developers (name) VALUES ('Soapbox');
INSERT INTO minis (name, image, slug, developer_id) VALUES ('Polls', '', '/polls', 1);
