package signature

import "os"

func init() {
	buffSize = 4 * os.Getpagesize()
}
