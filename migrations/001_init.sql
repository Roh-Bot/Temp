-- +goose Up
-- +goose StatementBegin

-- =========================
-- TABLES
-- =========================

CREATE TABLE users
(
    id UUID PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'admin')),

    CONSTRAINT users_username_unique UNIQUE (username),
    CONSTRAINT users_email_unique UNIQUE (email)
);


CREATE TABLE IF NOT EXISTS tasks
(
    id          uuid PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    status      VARCHAR(50)  NOT NULL CHECK (status IN ('pending', 'in_progress', 'completed')),
    user_id     uuid  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
    );

-- =========================
-- INDEXES
-- =========================

CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks (user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks (created_at);

-- =========================
-- SAMPLE DATA
-- =========================

INSERT INTO users (id, username, email, password, role)
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'admin', 'admin@example.com',
        '$2a$10$rZ8qH5YKzN5YvJxKGxKxOeYvXxXxXxXxXxXxXxXxXxXxXxXxXxXxX', 'admin'),
       ('550e8400-e29b-41d4-a716-446655440001', 'user1', 'user1@example.com',
        '$2a$10$rZ8qH5YKzN5YvJxKGxKxOeYvXxXxXxXxXxXxXxXxXxXxXxXxXxXxX', 'user')
    ON CONFLICT (username) DO NOTHING;

-- =========================
-- FUNCTIONS
-- =========================

CREATE OR REPLACE FUNCTION public.list_tasks_paginated(
    p_user_id uuid,
    p_is_admin BOOLEAN,
    p_status TEXT,
    p_limit INTEGER,
    p_last_id uuid DEFAULT NULL
)
RETURNS TABLE
(
    id          uuid,
    title       TEXT,
    description TEXT,
    status      TEXT,
    user_id     uuid,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    has_more    BOOLEAN,
    next_cursor uuid
)
LANGUAGE sql
AS
$$
WITH page AS (
    SELECT t.*
    FROM tasks t
    WHERE (p_is_admin OR t.user_id = p_user_id)
      AND (p_status IS NULL OR p_status = '' OR t.status = p_status)
      AND (p_last_id IS NULL OR t.id <= p_last_id)
    ORDER BY t.id DESC
    LIMIT p_limit + 1
)
SELECT
    id,
    title,
    description,
    status,
    user_id,
    created_at,
    updated_at,
    (SELECT COUNT(*) > p_limit FROM page) AS has_more,
    (SELECT id FROM page ORDER BY id DESC OFFSET p_limit LIMIT 1) AS next_cursor
FROM page
ORDER BY id DESC
    LIMIT p_limit;
$$;


CREATE OR REPLACE FUNCTION task_create(
    p_id uuid,
    p_title VARCHAR(255),
    p_description TEXT,
    p_status VARCHAR(50),
    p_user_id uuid,
    p_created_at TIMESTAMPTZ,
    p_updated_at TIMESTAMPTZ
)
    RETURNS VOID
    LANGUAGE plpgsql
AS
$$
BEGIN
INSERT INTO tasks (id, title, description, status, user_id, created_at, updated_at)
VALUES (p_id, p_title, p_description, p_status, p_user_id, p_created_at, p_updated_at);
END;
$$;

CREATE OR REPLACE FUNCTION task_get_by_id(
    p_id uuid,
    p_user_id uuid,
    p_is_admin BOOLEAN
)
    RETURNS TABLE
            (
                id          uuid,
                title       VARCHAR(255),
                description TEXT,
                status      VARCHAR(50),
                user_id     uuid,
                created_at  TIMESTAMPTZ,
                updated_at  TIMESTAMPTZ
            )
    LANGUAGE plpgsql
AS
$$
BEGIN
RETURN QUERY
SELECT t.id, t.title, t.description, t.status, t.user_id, t.created_at, t.updated_at
FROM tasks t
WHERE t.id = p_id
  AND (p_is_admin = TRUE OR t.user_id = p_user_id);
END;
$$;

CREATE OR REPLACE FUNCTION task_delete(
    p_id uuid,
    p_user_id uuid,
    p_is_admin BOOLEAN
)
    RETURNS VOID
    LANGUAGE plpgsql
AS
$$
BEGIN
DELETE
FROM tasks
WHERE id = p_id
  AND (p_is_admin = TRUE OR user_id = p_user_id);
END;
$$;

CREATE OR REPLACE FUNCTION task_update_status(
    p_id uuid,
    p_status VARCHAR(50)
)
    RETURNS VOID
    LANGUAGE plpgsql
AS
$$
BEGIN
UPDATE tasks
SET status     = p_status,
    updated_at = NOW()
WHERE id = p_id;
END;
$$;

CREATE OR REPLACE FUNCTION task_get_pending_older_than(
    p_minutes INT
)
    RETURNS TABLE
            (
                id          uuid,
                title       VARCHAR(255),
                description TEXT,
                status      VARCHAR(50),
                user_id     uuid,
                created_at  TIMESTAMPTZ,
                updated_at  TIMESTAMPTZ
            )
    LANGUAGE plpgsql
AS
$$
BEGIN
RETURN QUERY
SELECT t.id, t.title, t.description, t.status, t.user_id, t.created_at, t.updated_at
FROM tasks t
WHERE t.status IN ('pending', 'in_progress')
  AND t.created_at < NOW() - (p_minutes || ' minutes')::INTERVAL;
END;
$$;

CREATE OR REPLACE FUNCTION task_auto_complete_if_pending(
    p_task_id uuid
)
    RETURNS BOOLEAN
    LANGUAGE plpgsql
AS
$$
DECLARE
rows_updated INT;
BEGIN
UPDATE tasks
SET status     = 'completed',
    updated_at = NOW()
WHERE id = p_task_id
  AND status IN ('pending', 'in_progress');

GET DIAGNOSTICS rows_updated = ROW_COUNT;
RETURN rows_updated > 0;
END;
$$;

CREATE OR REPLACE FUNCTION user_create(
    p_id UUID,
    p_username VARCHAR,
    p_email VARCHAR,
    p_password VARCHAR,
    p_role VARCHAR
)
    RETURNS VOID
    LANGUAGE plpgsql
AS
$$
BEGIN
INSERT INTO users (id, username, email, password, role)
VALUES (p_id, p_username, p_email, p_password, p_role);
END;
$$;

CREATE OR REPLACE FUNCTION user_get_by_username(
    p_username VARCHAR
)
    RETURNS TABLE
            (
                id       UUID,
                username VARCHAR,
                email    VARCHAR,
                password VARCHAR,
                role     VARCHAR
            )
    LANGUAGE plpgsql
AS
$$
BEGIN
RETURN QUERY
SELECT u.id, u.username, u.email, u.password, u.role
FROM users u
WHERE u.username = p_username;

IF NOT FOUND THEN
        -- SQLSTATE 02000 (no_data_found)
        RAISE EXCEPTION 'User not found'
            USING ERRCODE = '02000';
END IF;
END;
$$;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP FUNCTION IF EXISTS task_auto_complete_if_pending;
DROP FUNCTION IF EXISTS task_get_pending_older_than;
DROP FUNCTION IF EXISTS task_update_status;
DROP FUNCTION IF EXISTS task_delete;
DROP FUNCTION IF EXISTS task_get_by_id;
DROP FUNCTION IF EXISTS task_create;
DROP FUNCTION IF EXISTS list_tasks_paginated;
DROP FUNCTION IF EXISTS user_create;
DROP FUNCTION IF EXISTS user_get_by_username;

DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_user_id;

DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
