-- +goose Up
CREATE TABLE uploaders (
    id text PRIMARY KEY,
    user_id text NOT NULL,
    algs text NOT NULL,
    description text NOT NULL,
    created_at datetime NOT NULL
);

-- +goose Down
DROP TABLE uploaders;
