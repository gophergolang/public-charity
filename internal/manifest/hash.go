// Adapted from tsdb's hashableUint64 + Label pattern:
// FNV-1a hashing for fast lookups and bloom filter keys.
package manifest

import "hash/fnv"

// HashID produces a stable FNV-1a uint64 hash for any string key.
// Used for: username lookups, category keys, interest matching, bloom filter insertion.
// Same hash family as tsdb's Label and MetricConfig.
func HashID(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// HashPair produces a combined hash for two strings (e.g., cell+category).
func HashPair(a, b string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(a))
	h.Write([]byte{0}) // separator
	h.Write([]byte(b))
	return h.Sum64()
}
