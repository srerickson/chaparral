// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package sqlite

import (
	"database/sql"
	"time"
)

type Object struct {
	ID      int64
	StoreID string
	OcflID  string
	Path    string
	Alg     string
	Spec    string
}

type ObjectContent struct {
	ObjectID int64
	Digest   string
	Paths    interface{}
	Fixity   interface{}
	Size     sql.NullInt64
}

type Upload struct {
	ID         string
	Size       int64
	UploaderID string
	Digests    []byte
}

type Uploader struct {
	ID          string
	UserID      string
	Algs        string
	Description string
	CreatedAt   time.Time
}

type Version struct {
	ObjectID    int64
	Num         int64
	State       interface{}
	Message     string
	UserName    sql.NullString
	UserAddress sql.NullString
	Created     time.Time
}
