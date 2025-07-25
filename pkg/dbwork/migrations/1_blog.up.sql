CREATE TABLE users(
  id BIGSERIAL PRIMARY KEY,
  login TEXT,
  password VARCHAR
);

CREATE TABLE articles(
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT,
  text TEXT
);
