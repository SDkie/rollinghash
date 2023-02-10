package rabinkarp

// The RabinKarp seed value.
//
// The seed ensures different length zero blocks have different hashes. It
// effectively encodes the length into the hash.
const RABINKARP_SEED uint32 = 1

// The RabinKarp multiplier.
//
// This multiplier has a bit pattern of 1's getting sparser with significance,
// is the product of 2 large primes, and matches the characterstics for a good
// LCG multiplier.
const RABINKARP_MULT uint32 = 135283237

// The RabinKarp inverse multiplier.
//
// This is the inverse of RABINKARP_MULT modular 2^32. Multiplying by this is
// equivalent to dividing by RABINKARP_MULT.
const RABINKARP_INVM uint32 = 2565867949

// The RabinKarp seed adjustment.
//
// This is a factor used to adjust for the seed when rolling out values. It's
// equal to; (RABINKARP_MULT - 1) * RABINKARP_SEED
const RABINKARP_ADJ uint32 = 135283236

// Hash returns the hash and the appropriate multiplicative
// factor for use in Rabin-Karp algorithm.
func Hash(sep []byte) (uint32, uint32) {
	hash := RABINKARP_SEED
	for i := 0; i < len(sep); i++ {
		hash = hash*RABINKARP_MULT + uint32(sep[i])
	}
	var pow, sq uint32 = 1, RABINKARP_MULT
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

// Rotate create the new hash of the rotated chunk
func Rotate(hash, pow, old, new uint32) uint32 {
	return hash*RABINKARP_MULT - (old+RABINKARP_ADJ)*pow + new
}

// RollOut creates new hash when we rollout one character
func RollOut(hash, pow, old uint32) (uint32, uint32) {
	pow *= RABINKARP_INVM
	hash -= pow * (old + RABINKARP_ADJ)
	return hash, pow
}
