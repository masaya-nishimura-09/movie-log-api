-- =====================
-- password is "Password1" (BCrypt)
-- =====================

-- users
INSERT INTO
    users (
        id,
        username,
        email,
        password,
        created_at,
        updated_at
    )
VALUES
    (
        1,
        'Test User',
        'test@example.com',
        '{bcrypt}$2b$10$u9X60onSaZLwlhp1wAhnzeWPW2bP3qaS0iLjjcP2USk8ZCOLwacuy',
        NOW (),
        NOW ()
    ),
    (
        2,
        'Test User 2',
        'test2@example.com',
        '{bcrypt}$2b$10$u9X60onSaZLwlhp1wAhnzeWPW2bP3qaS0iLjjcP2USk8ZCOLwacuy',
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
