-- Create users table
CREATE TABLE IF NOT EXISTS users
(
    id       VARCHAR(36) PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email    VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255)        NOT NULL,
    role     VARCHAR(50)         NOT NULL CHECK (role IN ('user', 'admin'))
);

-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks
(
    id          VARCHAR(36) PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    status      VARCHAR(50)  NOT NULL CHECK (status IN ('pending', 'in_progress', 'completed')),
    user_id     VARCHAR(36)  NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks (user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks (created_at);

-- Insert sample users (password is 'password123' hashed with bcrypt)
INSERT INTO users (id, username, email, password, role)
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'admin', 'admin@example.com',
        '$2a$10$rZ8qH5YKzN5YvJxKGxKxOeYvXxXxXxXxXxXxXxXxXxXxXxXxXxXxX', 'admin'),
       ('550e8400-e29b-41d4-a716-446655440001', 'user1', 'user1@example.com',
        '$2a$10$rZ8qH5YKzN5YvJxKGxKxOeYvXxXxXxXxXxXxXxXxXxXxXxXxXxXxX', 'user')
ON CONFLICT (username) DO NOTHING;

-- Stored procedure for listing tasks with pagination
CREATE OR REPLACE FUNCTION list_tasks_paginated(
    p_user_id VARCHAR(36),
    p_is_admin BOOLEAN,
    p_status VARCHAR(50),
    p_limit INT,
    p_scroll_token TEXT
)
    RETURNS TABLE
            (
                id               VARCHAR(36),
                title            VARCHAR(255),
                description      TEXT,
                status           VARCHAR(50),
                user_id          VARCHAR(36),
                created_at       TIMESTAMP,
                updated_at       TIMESTAMP,
                total_count      BIGINT,
                next_scroll_token TEXT
            )
AS
$$
DECLARE
    v_total_count     BIGINT;
    v_last_created_at TIMESTAMP;
    v_last_id         VARCHAR(36);
BEGIN
    -- Decode scroll token (format: created_at|id)
    IF p_scroll_token IS NOT NULL AND p_scroll_token <> '' THEN
        v_last_created_at := split_part(p_scroll_token, '|', 1)::timestamp;
        v_last_id := split_part(p_scroll_token, '|', 2);
    END IF;

    -- Total count
    SELECT COUNT(*)
    INTO v_total_count
    FROM tasks t
    WHERE (
        p_is_admin = TRUE
            OR t.user_id = p_user_id
        )
      AND (p_status IS NULL OR p_status = '' OR t.status = p_status);

    -- Return page
    RETURN QUERY
        WITH page AS (
            SELECT t.*
            FROM tasks t
            WHERE (
                p_is_admin = TRUE
                    OR t.user_id = p_user_id
                )
              AND (p_status IS NULL OR p_status = '' OR t.status = p_status)
              AND (
                v_last_created_at IS NULL
                    OR (t.created_at, t.id) < (v_last_created_at, v_last_id)
                )
            ORDER BY t.created_at DESC, t.id DESC
            LIMIT p_limit
        )
        SELECT
            p.id,
            p.title,
            p.description,
            p.status,
            p.user_id,
            p.created_at,
            p.updated_at,
            v_total_count,
            CASE
                WHEN COUNT(*) OVER () = p_limit
                    THEN
                    (
                        SELECT
                            (last_row.created_at || '|' || last_row.id)
                        FROM (
                                 SELECT created_at, id
                                 FROM page
                                 ORDER BY created_at DESC, id DESC
                                 OFFSET p_limit - 1 LIMIT 1
                             ) last_row
                    )
                END AS next_scroll_token
        FROM page p;

END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION task_create(
    p_id VARCHAR(36),
    p_title VARCHAR(255),
    p_description TEXT,
    p_status VARCHAR(50),
    p_user_id VARCHAR(36),
    p_created_at TIMESTAMP,
    p_updated_at TIMESTAMP
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
    p_id VARCHAR(36),
    p_user_id VARCHAR(36),
    p_is_admin BOOLEAN
)
    RETURNS TABLE
            (
                id VARCHAR(36),
                title VARCHAR(255),
                description TEXT,
                status VARCHAR(50),
                user_id VARCHAR(36),
                created_at TIMESTAMP,
                updated_at TIMESTAMP
            )
    LANGUAGE plpgsql
AS
$$
BEGIN
    RETURN QUERY
        SELECT t.id,
               t.title,
               t.description,
               t.status,
               t.user_id,
               t.created_at,
               t.updated_at
        FROM tasks t
        WHERE t.id = p_id
          AND (
            p_is_admin = TRUE
                OR t.user_id = p_user_id
            );
END;
$$;

CREATE OR REPLACE FUNCTION task_delete(
    p_id VARCHAR(36),
    p_user_id VARCHAR(36),
    p_is_admin BOOLEAN
)
    RETURNS VOID
    LANGUAGE plpgsql
AS
$$
BEGIN
    DELETE FROM tasks
    WHERE id = p_id
      AND (
        p_is_admin = TRUE
            OR user_id = p_user_id
        );
END;
$$;

CREATE OR REPLACE FUNCTION task_update_status(
    p_id VARCHAR(36),
    p_status VARCHAR(50)
)
    RETURNS VOID
    LANGUAGE plpgsql
AS
$$
BEGIN
    UPDATE tasks
    SET status = p_status,
        updated_at = NOW()
    WHERE id = p_id;
END;
$$;

CREATE OR REPLACE FUNCTION task_get_pending_older_than(
    p_minutes INT
)
    RETURNS TABLE
            (
                id VARCHAR(36),
                title VARCHAR(255),
                description TEXT,
                status VARCHAR(50),
                user_id VARCHAR(36),
                created_at TIMESTAMP,
                updated_at TIMESTAMP
            )
    LANGUAGE plpgsql
AS
$$
BEGIN
    RETURN QUERY
        SELECT t.id,
               t.title,
               t.description,
               t.status,
               t.user_id,
               t.created_at,
               t.updated_at
        FROM tasks t
        WHERE t.status IN ('pending', 'in_progress')
          AND t.created_at < NOW() - (p_minutes || ' minutes')::interval;
END;
$$;

CREATE OR REPLACE FUNCTION task_auto_complete_if_pending(
    p_task_id VARCHAR
)
    RETURNS BOOLEAN
    LANGUAGE plpgsql
AS $$
DECLARE
    rows_updated INT;
BEGIN
    UPDATE tasks
    SET status = 'completed',
        updated_at = NOW()
    WHERE id = p_task_id
      AND status IN ('pending', 'in_progress');

    GET DIAGNOSTICS rows_updated = ROW_COUNT;

    RETURN rows_updated > 0;
END;
$$;
