package main

import (
	"math/rand"
)

// v1CodeAlphabet is the Crockford-ish alphabet used for room codes.
// 30 characters: digits 2-9 and uppercase A-Z minus I, L, O, U.
// Chosen so codes read aloud are unambiguous.
const v1CodeAlphabet = "23456789ABCDEFGHJKMNPQRSTVWXYZ"

// v1CodeLen is the length of generated room codes.
const v1CodeLen = 6

// GenerateV1RoomCode returns a fresh random 6-character room code drawn
// from v1CodeAlphabet. With 30^6 ≈ 729M combinations, callers can safely
// retry on collision.
func GenerateV1RoomCode() string {
	b := make([]byte, v1CodeLen)
	for i := range b {
		b[i] = v1CodeAlphabet[rand.Intn(len(v1CodeAlphabet))]
	}
	return string(b)
}
