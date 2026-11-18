-- =====================
-- password is "Password1" (BCrypt)
-- =====================

-- users
INSERT INTO
    users (
        id,
        username,
        email,
        hashed_password,
        role,
        email_verified_at,
        created_at,
        updated_at
    )
VALUES
    (
        1,
        'Test User',
        'test@example.com',
        '$2b$10$u9X60onSaZLwlhp1wAhnzeWPW2bP3qaS0iLjjcP2USk8ZCOLwacuy',
        'user',
        NOW (),
        NOW (),
        NOW ()
    ),
    (
        2,
        'Test User 2',
        'test2@example.com',
        '$2b$10$u9X60onSaZLwlhp1wAhnzeWPW2bP3qaS0iLjjcP2USk8ZCOLwacuy',
        'user',
        NOW (),
        NOW (),
        NOW ()
    );

-- reset sequence
SELECT
    SETVAL (
        'users_id_seq',
        (
            SELECT
                MAX(id)
            FROM
                users
        )
    );
