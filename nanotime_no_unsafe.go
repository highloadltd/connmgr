// +build appengine windows

package connmgr

import "time"

func nanotime() int64 {
	return time.Now().UnixNano()
}
