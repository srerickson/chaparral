-- name: CreateUploader :one
INSERT INTO uploaders (
    id, 
    user_id,
    algs,
    description,
    created_at
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING *;


-- name: GetUploader :one
SELECT * FROM uploaders WHERE id = ? LIMIT 1;

-- name: GetUploaderIDs :many
SELECT id FROM uploaders ORDER BY created_at;

-- name: CountUploaders :one
SELECT COUNT(*) FROM uploaders;

-- name: DeleteUploader :exec
DELETE FROM uploaders WHERE id = ?;

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
SELECT * FROM uploads WHERE uploader_id = ?;

-- name: DeleteUploads :exec
DELETE FROM uploads WHERE uploader_id = ?;


-- name: GetObject :one
SELECT * FROM objects WHERE store_id = ? AND ocfl_id = ?;

-- name: CreateObject :one
INSERT INTO objects (
    store_id,
    ocfl_id,
    path,
    spec,
    alg
) VALUES (?1, ?2, ?3, ?4, ?5)
ON CONFLICT(store_id, ocfl_id) DO UPDATE SET
    path=?3,
    spec=?4,
    alg=?5
RETURNING *;

-- name: DeleteObject :exec
DELETE FROM objects WHERE store_id = ? AND ocfl_id = ?;

-- name: GetObjectContent :one
SELECT * FROM object_contents WHERE object_id = ? AND digest = ?;

-- name: GetObjectContents :many
SELECT * FROM object_contents WHERE object_id = ?;

-- name: CreateObjectContent :one 
INSERT INTO object_contents (
    object_id,
    digest,
    paths,
    fixity,
    size
) VALUES (?1, ?2, ?3, ?4, ?5)
ON CONFLICT(object_id, digest) DO UPDATE SET
    paths=?3,
    fixity=?4,
    size=?5
RETURNING *;    

-- name: DeleteObjectContents :exec
DELETE FROM object_contents WHERE object_id = (
    SELECT id FROM objects WHERE store_id = ? AND ocfl_id = ?
);
