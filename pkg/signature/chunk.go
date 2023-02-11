package signature

import (
	"math"
)

// getOptimalChunkSize returns the optimal chunk size for a given file size.
// The optimal chunk size is sqrt(filesize) with a 256 min size rounded down to a multiple of 128.
func getOptimalChunkSize(filesize int64) uint32 {
	chunkLen := 256
	if filesize > 256*256 {
		chunkLen = int(math.Sqrt(float64(filesize))) & -128
	}

	return uint32(chunkLen)
}
