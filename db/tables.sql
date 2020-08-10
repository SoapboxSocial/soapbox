CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  display_name VARCHAR(256) NOT NULL,
  username VARCHAR(100) NOT NULL,
  email VARCHAR(254) NOT NULL
);

CREATE UNIQUE INDEX idx_email ON users (email);
CREATE UNIQUE INDEX idx_username ON users (username);

CREATE TABLE IF NOT EXISTS followers (
    follower INT NOT NULL,
    user_id INT NOT NULL,
    FOREIGN KEY (follower) REFERENCES users(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    PRIMARY KEY (follower, user_id)
);
