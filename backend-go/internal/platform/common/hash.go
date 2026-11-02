package common

import "unicode/utf16"

// JavaStringHashCode reproduces java.lang.String.hashCode() exactly: h = 31*h + c over each UTF-16
// code unit, as a 32-bit signed int (with overflow). Kotlin's String.hashCode() is the same. Used to
// match the Kotlin backend's content fingerprints (member/promotion) byte-for-byte in the snapshot
// version.
func JavaStringHashCode(s string) int32 {
	var h int32
	for _, u := range utf16.Encode([]rune(s)) {
		h = 31*h + int32(u)
	}
	return h
}
