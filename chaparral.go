package chaparral

import "runtime/debug"

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
