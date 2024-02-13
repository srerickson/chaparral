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

type ObjectVersion struct {
	Version         int
	Spec            string
	Head            int
	DigestAlgorithm string
	State           Manifest
	Messsage        string
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
	Size   int64
	Paths  []string
	Fixity ocfl.DigestSet
}
