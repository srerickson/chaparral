package chaparral

import (
	"runtime/debug"
	"time"

	chapv1connect "github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"github.com/srerickson/ocfl-go"
)

var VERSION = "devel"

var CODE_VERSION = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		revision := ""
		// revtime := ""
		localmods := false
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				revision = setting.Value
			// case "vcs.time":
			// 	revtime = setting.Value
			case "vcs.modified":
				localmods = setting.Value == "true"
			}
		}
		if !localmods {
			return revision
		}
	}
	return "none"
}()

// http routes for upload/download
const (
	RouteDownload = "/" + chapv1connect.AccessServiceName + "/" + "download"
	RouteUpload   = `/` + chapv1connect.CommitServiceName + "/" + "upload"

	QueryDigest      = "digest"
	QueryContentPath = "content_path"
	QueryObjectID    = "object_id"
	QueryUploaderID  = "uploader"
	QueryStorageRoot = "storage_root"
)

// ObjectManifest corresponds to GetObjectManifestResponse proto
type ObjectManifest struct {
	ObjectID        string
	StorageRootID   string
	Path            string
	Spec            string
	DigestAlgorithm string
	Manifest        Manifest
}

// ObjectVersion corresponds to GetObjectVersionResponse proto
type ObjectVersion struct {
	ObjectID        string
	StorageRootID   string
	Spec            string
	Version         int
	Head            int
	DigestAlgorithm string
	State           Manifest
	Message         string
	User            *ocfl.User
	Created         time.Time
}

// Manifest maps digest to FileInfo
type Manifest map[string]FileInfo

func (m Manifest) DigestMap() ocfl.DigestMap {
	dm := map[string][]string{}
	for d, info := range m {
		dm[d] = info.Paths
	}
	return ocfl.DigestMap(dm)
}

func (m Manifest) PathMap() ocfl.PathMap {
	pm := map[string]string{}
	for d, info := range m {
		for _, p := range info.Paths {
			pm[p] = d
		}
	}
	return pm
}

type FileInfo struct {
	Size   int64          // number of bytes
	Paths  []string       // sorted slice of path names
	Fixity ocfl.DigestSet // other digests associated witht the content
}

type Uploader struct {
	ID               string    `json:"id"`
	UploadPath       string    `json:"upload_path"`
	DigestAlgorithms []string  `json:"digest_algorithms"`
	Description      string    `json:"description"`
	Created          time.Time `json:"created"`
	UserID           string    `json:"user_id"`
	Uploads          []Upload  `json:"uploads,omitempty"`
}

type Upload struct {
	Size    int64          `json:"size"`
	Digests ocfl.DigestSet `json:"digests"`
}

// UploadResult is response from call to upload
type UploadResult struct {
	Upload
	Err string `json:"err"`
}
