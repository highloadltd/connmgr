// +build !appengine,!windows

package connmgr

import (
	_ "unsafe" // required to use //go:linkname
)

// Nanotime is monotonic time provider.
//go:noescape
//go:linkname nanotime runtime.nanotime
func nanotime() int64
