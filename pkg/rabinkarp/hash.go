package rabinkarp

// PrimeRK is the prime base used in Rabin-Karp algorithm.
const PrimeRK = 16777619

// Hash returns the hash and the appropriate multiplicative
// factor for use in Rabin-Karp algorithm.
func Hash(sep []byte) (uint32, uint32) {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*PrimeRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, PrimeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

func RollingHash(hash, pow, old, new uint32) uint32 {
	return hash*PrimeRK - uint32(old)*pow + uint32(new)
}
