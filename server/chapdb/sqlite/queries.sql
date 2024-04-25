-- 
-- Uploaders
--

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

--
-- Objects / ObjectContents
--

-- name: GetObject :one
SELECT * FROM objects WHERE store_id = ? AND ocfl_id = ?;

-- name: AllObjects :many
SELECT * FROM objects WHERE store_id = ? ORDER BY ocfl_id ASC;

-- name: CountObjects :one
SELECT COUNT(*) FROM objects WHERE store_id = ?;

-- name: ListObjects :many
SELECT * FROM objects WHERE store_id = ?1 AND ocfl_id > ?2
    ORDER BY ocfl_id ASC LIMIT ?3;

-- name: CreateObject :one
INSERT INTO objects (
    store_id,
    ocfl_id,
    path,
    spec,
    alg,
    indexed_at
) VALUES (?1, ?2, ?3, ?4, ?5, ?6)
ON CONFLICT(store_id, ocfl_id) DO UPDATE SET
    path=?3,
    spec=?4,
    alg=?5,
    indexed_at=?6
RETURNING *;

-- name: DeleteObject :exec
DELETE FROM objects WHERE store_id = ? AND ocfl_id = ?;

-- name: DeleteStaleObjects :exec
DELETE FROM objects WHERE store_id = ? AND indexed_at < ?;

-- name: GetObjectContent :one
SELECT * FROM object_contents WHERE object_id = ? AND digest = ? LIMIT 1;

-- name: GetObjectContents :many
SELECT * FROM object_contents WHERE object_id = ?
    ORDER BY digest ASC;

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
DELETE FROM object_contents WHERE object_id IN (
    SELECT id FROM objects WHERE store_id = ? AND ocfl_id = ?
);

-- name: DeleteOrphanedObjectContents :exec
DELETE FROM object_contents WHERE object_id IN (
    SELECT object_contents.object_id
    FROM object_contents
    LEFT JOIN objects
    ON objects.id = object_contents.object_id
    WHERE objects.id is NULL
);
