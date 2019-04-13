package errorx

import "github.com/vsaien/cuter/lib/logx"

func Rescue(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		logx.Severe(p)
	}
}
