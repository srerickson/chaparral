
-- +goose Up
CREATE TABLE objects (
    id INTEGER PRIMARY KEY,  -- internal db ID
    store_id TEXT NOT NULL, -- storage root ID
    ocfl_id TEXT NOT NULL, -- object id from inventory
    path TEXT NOT NULL, -- object path relative to FS
    alg TEXT NOT NULL, -- object's primary digest algorithm
    spec TEXT NOT NULL, -- OCFL spec
    UNIQUE(store_id, path),
    UNIQUE(store_id, ocfl_id)
);

CREATE TABLE versions (
    object_id INTEGER NOT NULL, -- objects table FK
    num INTEGER NOT NULL, -- version num (1,2,3,...)
    state BLOB NOT NULL, -- json object: {digest: [path1, path2]}
    message TEXT NOT NULL, -- message saved with object version
    user_name TEXT, -- user name saved with object version
    user_address TEXT, -- user address saved with object version
    created DATETIME NOT NULL, -- create timestamp for object version
    
    PRIMARY KEY(object_id, num)
);

CREATE TABLE object_contents (
    object_id INTEGER NOT NULL, -- objects table FK
    digest TEXT NOT NULL, -- digest bytes for the object_contents
    paths BLOB NOT NULL, -- array of paths
    fixity BLOB NOT NULL, -- json object with alternate digests {alg: digest}
    size INTEGER NOT NULL, -- may not be available
    
    PRIMARY KEY(object_id, digest)
);

-- +goose Down
DROP TABLE objects;
DROP TABLE versions;
DROP TABLE object_contents;
