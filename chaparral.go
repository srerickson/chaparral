package chaparral

import (
	"github.com/srerickson/ocfl-go"
	"runtime/debug"
	"time"
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

type ObjectState struct {
	StorageRootID   string
	ObjectID        string
	Spec            string
	Version         int
	Head            int
	DigestAlgorithm string
	State           map[string]FileInfo
	Messsage        string
	User            *ocfl.User
	Created         time.Time
}

func (obj ObjectState) DigestMap() ocfl.DigestMap {
	m := map[string][]string{}
	for d, info := range obj.State {
		m[d] = info.Paths
	}
	mp, err := ocfl.NewDigestMap(m)
	if err != nil {
		panic(err)
	}
	return mp
}

type ObjectManifest struct {
	StorageRootID   string
	ObjectID        string
	DigestAlgorithm string
	Manifest        map[string]FileInfo
}

type FileInfo struct {
	Size   int64
	Paths  []string
	Fixity map[string]string
}
