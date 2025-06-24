CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(96) NOT NULL UNIQUE,
    level INT NOT NULL DEFAULT 0,
    description TEXT
);

INSERT INTO
    roles (name, description, level)
VALUES (
        'user',
        'A user can create posts and comments',
        1
);

INSERT INTO
    roles (name, description, level)
VALUES (
           'moderator',
           'A moderator can edit posts of other users',
           2
);

INSERT INTO
    roles (name, description, level)
VALUES (
           'admin',
           'Administrator',
           3
);