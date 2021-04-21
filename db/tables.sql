SET timezone = 'Europe/Zurich';

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

CREATE TABLE IF NOT EXISTS mini_developers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE UNIQUE INDEX idx_mini_developers_name ON mini_developers (name);

CREATE TABLE IF NOT EXISTS current_rooms (
    user_id INT NOT NULL,
    room VARCHAR(27) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_current_rooms ON current_rooms (room);
CREATE UNIQUE INDEX idx_current_rooms_user_id ON current_rooms (user_id, room);
CREATE UNIQUE INDEX idx_current_rooms_user ON current_rooms (user_id);

CREATE OR REPLACE FUNCTION update_current_rooms(id INT, room_id VARCHAR(27))
    RETURNS VOID
    AS $current_rooms$
    BEGIN
        -- This inserts or updates the room id a user is in.
        INSERT INTO current_rooms(user_id, room)
        VALUES(id, room_id)
            ON CONFLICT (user_id)
            DO
                UPDATE SET room = room_id;
    END;
    $current_rooms$
    LANGUAGE PLPGSQL;

CREATE TABLE IF NOT EXISTS minis (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    image VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    size INT NOT NULL DEFAULT 1, -- 0 - Small, 1 - Regular, 2 - large
    description TEXT NOT NULL,
    developer_id INT NOT NULL,
    FOREIGN KEY (developer_id) REFERENCES mini_developers(id) ON DELETE CASCADE
);

-- Inserting apps
INSERT INTO mini_developers (name) VALUES ('Soapbox');
INSERT INTO minis (name, image, slug, size, developer_id) VALUES ('Polls', '', '/polls', 1, 1);

CREATE TABLE IF NOT EXISTS user_room_logs (
    user_id INT NOT NULL,
    room VARCHAR(27) NOT NULL,
    join_time TIMESTAMPTZ,
    left_time TIMESTAMPTZ,
    visibility VARCHAR(7),
    CHECK (visibility IN ('public', 'private')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_user_room_logs_user_join ON user_room_logs (user_id, room, join_time);

CREATE TABLE IF NOT EXISTS user_room_time (
    user_id INT NOT NULL,
    seconds INT NOT NULL,
    visibility VARCHAR(7),
    CHECK (visibility IN ('public', 'private')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_user_room_time ON user_room_time (user_id, visibility);

CREATE OR REPLACE FUNCTION update_user_room_time()
    RETURNS TRIGGER
    AS $user_room_time$
    BEGIN
        -- This inserts or updates the time a user has spent in a specific room type.
        INSERT INTO user_room_time(user_id, seconds, visibility)
        VALUES(NEW.user_id, EXTRACT(EPOCH FROM (NEW.left_time - NEW.join_time)), NEW.visibility)
            ON CONFLICT (user_id, visibility)
            DO
                UPDATE SET seconds = user_room_time.seconds + EXTRACT(EPOCH FROM (NEW.left_time - NEW.join_time));
        RETURN NEW;
    END;
    $user_room_time$
    LANGUAGE PLPGSQL;

CREATE TRIGGER calculate_room_time
    AFTER INSERT
    ON user_room_logs
    FOR EACH ROW
    EXECUTE PROCEDURE update_user_room_time();

CREATE TABLE IF NOT EXISTS user_active_times (
    user_id INT NOT NULL,
    last_active TIMESTAMPTZ,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_user_active_times ON user_active_times (user_id);

CREATE OR REPLACE FUNCTION update_user_active_times(id INT, active TIMESTAMPTZ)
    RETURNS VOID
    AS $user_active_times$
    BEGIN
        -- This inserts or updates the time a user has spent in a specific room type.
        INSERT INTO user_active_times(user_id, last_active)
        VALUES(id, active)
            ON CONFLICT (user_id)
            DO
                UPDATE SET last_active = active;
    END;
    $user_active_times$
    LANGUAGE PLPGSQL;

-- @TODO BETTER NAME
CREATE TABLE IF NOT EXISTS notification_subscriptions (
    subscriber INT NOT NULL,
    user_id INT NOT NULL,
    FOREIGN KEY (subscriber) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_notification_subscriptions ON notification_subscriptions (subscriber, user_id);
