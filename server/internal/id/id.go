// Package id generates opaque, random resource IDs, shared by every
// feature's repository so ID generation is written once.
package id

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns a 32-character hex-encoded random ID.
func New() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
