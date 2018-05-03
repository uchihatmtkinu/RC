package basic

import (
	"crypto/sha256"
)

//DoubleHash256 returns the double hash result of hash256
func DoubleHash256(a *[]byte, b *[32]byte) {
	tmp := sha256.Sum256(*a)
	*b = sha256.Sum256(tmp[:])
}
