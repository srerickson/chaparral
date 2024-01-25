
-- +goose Up
CREATE TABLE objects (
    id INTEGER PRIMARY KEY,  -- internal db ID
    store_id TEXT NOT NULL, -- storage root ID
    ocfl_id TEXT NOT NULL, -- object id from inventory
    path TEXT NOT NULL, -- object path relative to storage root
    head INTEGER NOT NULL, -- current version
    spec TEXT NOT NULL, -- OCFL spec
    alg TEXT NOT NULL, -- digest algorithm
    UNIQUE(store_id, path),
    UNIQUE(store_id, ocfl_id)
);

CREATE TABLE versions (
    object_id INTEGER NOT NULL, -- objects table FK
    num INTEGER NOT NULL, -- version num (1,2,3,...)
    state BLOB NOT NULL, -- json object: {path: digest}
    message TEXT NOT NULL, -- message saved with object version
    user_name TEXT, -- user name saved with object version
    user_address TEXT, -- user address saved with object version
    created DATETIME NOT NULL, -- create timestamp for object version
    
    PRIMARY KEY(object_id, num)
);

CREATE TABLE content (
    object_id INTEGER NOT NULL, -- objects table FK
    digest BLOB NOT NULL, -- digest bytes for the content
    path TEXT NOT NULL, -- path to content
    fixity BLOB, -- json object with alternate digests {alg: digest}
    size INTEGER, -- may not be available
    
    PRIMARY KEY(object_id, digest)
);

-- +goose Down
DROP TABLE objects;
DROP TABLE versions;
DROP TABLE content;
