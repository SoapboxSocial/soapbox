SET timezone = 'Europe/Zurich';

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    display_name VARCHAR(256) NOT NULL,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(254) NOT NULL,
    image VARCHAR(100) NOT NULL DEFAULT '',
    bio TEXT NOT NULL DEFAULT '',
    joined TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
    weight INT NOT NULL DEFAULT 0,
    FOREIGN KEY (developer_id) REFERENCES mini_developers(id) ON DELETE CASCADE
);

-- Inserting apps
INSERT INTO mini_developers (name) VALUES ('Soapbox');
INSERT INTO minis (name, image, slug, size, developer_id) VALUES ('Polls', '', '/polls', 1, 1);

CREATE TABLE IF NOT EXISTS mini_scores (
    room VARCHAR(27) NOT NULL,
    mini_id INT NOT NULL,
    user_id INT NOT NULL,
    time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    score INT NOT NULL,
    FOREIGN KEY (mini_id) REFERENCES minis(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_mini_scores ON mini_scores (user_id, mini_id, room, time);

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

CREATE TABLE IF NOT EXISTS notification_settings (
    user_id INT NOT NULL,
    room_frequency INT NOT NULL DEFAULT 2,
    follows BOOLEAN NOT NULL DEFAULT true,
    welcome_rooms BOOLEAN NOT NULL DEFAULT true,
    CHECK (room_frequency IN (0, 1, 2, 3)), -- 0 = off, 1 - infrequent, 2 - normal, 3 - frequent
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION insert_notification_settings() RETURNS TRIGGER AS
    $notification_settings$
    BEGIN
        INSERT INTO notification_settings(user_id) VALUES(new.id);
        RETURN new;
    END;
    $notification_settings$
    language plpgsql;

CREATE TRIGGER insert_notification_settings_trigger
    AFTER INSERT ON users
    FOR EACH ROW
    EXECUTE PROCEDURE insert_notification_settings();

CREATE TABLE IF NOT EXISTS notification_analytics (
    id VARCHAR(36) NOT NULL,
    target INT NOT NULL,
    origin INT,
    category TEXT NOT NULL,
    sent TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    opened TIMESTAMPTZ,
    room VARCHAR(27),
    FOREIGN KEY (target) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (origin) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_notification_analytics ON notification_analytics (id, target);

CREATE TABLE IF NOT EXISTS follow_recommendations (
    user_id INT NOT NULL,
    recommendation INT NOT NULL,
    recommended TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (recommendation) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_follow_recommendations ON follow_recommendations (user_id, recommendation);

CREATE TABLE IF NOT EXISTS last_follow_recommended (
    user_id INT NOT NULL,
    last_recommended TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_last_follow_recommended ON last_follow_recommended (user_id);

CREATE OR REPLACE FUNCTION insert_last_follow_recommended() RETURNS TRIGGER AS
    $last_follow_recommended$
    BEGIN
        INSERT INTO last_follow_recommended(user_id) VALUES(new.id);
        RETURN new;
    END;
    $last_follow_recommended$
    language plpgsql;

CREATE TRIGGER insert_last_follow_recommended_trigger
    AFTER INSERT ON users
    FOR EACH ROW
    EXECUTE PROCEDURE insert_last_follow_recommended();

CREATE OR REPLACE FUNCTION delete_follow_recommendations() RETURNS TRIGGER AS
    $follow_recommendations$
    BEGIN
        DELETE FROM follow_recommendations WHERE user_id = new.follower AND recommendation = new.user_id;
        RETURN new;
    END;
    $follow_recommendations$
language plpgsql;

CREATE TRIGGER delete_follow_recommendations_trigger
    AFTER INSERT ON followers
    FOR EACH ROW
    EXECUTE PROCEDURE delete_follow_recommendations();
