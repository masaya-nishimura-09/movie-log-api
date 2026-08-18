CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
);

CREATE TABLE refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    revoked_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    release_year INTEGER NOT NULL,
    runtime INTEGER NOT NULL,
    language TEXT NOT NULL,
    poster_url TEXT NOT NULL,
    watched_at TIMESTAMPTZ NOT NULL,
    platform TEXT NOT NULL,
    score INTEGER NOT NULL,
    memo TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
);

CREATE INDEX records_user_id_watched_at_idx ON records (user_id, watched_at DESC, id DESC);

CREATE TABLE record_genres (
    id BIGSERIAL PRIMARY KEY,
    record_id BIGINT NOT NULL REFERENCES records (id) ON DELETE CASCADE,
    value TEXT NOT NULL,
    UNIQUE (record_id, value)
);

CREATE TABLE record_countries (
    id BIGSERIAL PRIMARY KEY,
    record_id BIGINT NOT NULL REFERENCES records (id) ON DELETE CASCADE,
    value TEXT NOT NULL,
    UNIQUE (record_id, value)
);

CREATE TABLE record_mood_tags (
    id BIGSERIAL PRIMARY KEY,
    record_id BIGINT NOT NULL REFERENCES records (id) ON DELETE CASCADE,
    value TEXT NOT NULL,
    UNIQUE (record_id, value)
);

CREATE TABLE record_credits (
    id BIGSERIAL PRIMARY KEY,
    record_id BIGINT NOT NULL REFERENCES records (id) ON DELETE CASCADE,
    person_name TEXT NOT NULL,
    credit_role TEXT NOT NULL
);

CREATE INDEX record_credits_record_id_idx ON record_credits (record_id);
