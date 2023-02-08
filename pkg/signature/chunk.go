package signature

import "math"

// getOptimalChunkSize returns the optimal chunk size for a given file size.
// The optimal chunk size is sqrt(filesize) with a 256 min size rounded down to a multiple of 128.
func getOptimalChunkSize(filesize int64) int64 {
	chunkLen := int64(256)
	if filesize > 256*256 {
		chunkLen = int64(math.Sqrt(float64(filesize))) & -128
	}

	return chunkLen
}
