CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  display_name VARCHAR(256) NOT NULL,
  username VARCHAR(100) NOT NULL,
  email VARCHAR(254) NOT NULL
);

CREATE INDEX idx_email ON users (email);
