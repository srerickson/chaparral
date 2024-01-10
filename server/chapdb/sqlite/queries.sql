-- name: CreateUploader :one
INSERT INTO uploaders (
    id, 
    root_id,
    user_id,
    algs,
    description,
    created_at
) VALUES (
    ?, ?, ?, ?, ?, ?
) RETURNING *;


-- name: GetUploader :one
SELECT * from uploaders where id = ? LIMIT 1;

-- name: GetUploaderIDs :many
SELECT id from uploaders ORDER BY created_at;

-- name: CountUploaders :one
SELECT COUNT(*) from uploaders;

-- name: DeleteUploader :exec
DELETE from uploaders where id = ?;

-- name: CreateUpload :one
INSERT INTO uploads (
    id, 
    uploader_id,
    size,
    digests
) VALUES (
    ?, ?, ?, ?
) RETURNING *;


-- name: GetUploads :many
SELECT * from uploads WHERE uploader_id = ?;

-- name: DeleteUploads :exec
DELETE from uploads where uploader_id = ?;