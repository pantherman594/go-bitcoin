package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const WIDTH = 4 // 256 (sha256) / 64 (uint64) = 4

// Stores a 256-bit number in uint64 numbers.
type uint256 [WIDTH]uint64

// Zeros out the entire uint256
func (u *uint256) Zero() {
	for i := 0; i < WIDTH; i++ {
		u[i] = 0
	}
}

func (u *uint256) FromUint64(o uint64) {
	u.Zero()
	u[0] = o
}

// Convert an array of 32 bytes into an array of 4 uint64.
func (u *uint256) FromBytes(b [sha256.Size]byte) {
	u.Zero()
	for i := 0; i < WIDTH; i++ {
		u[WIDTH-i-1] = binary.BigEndian.Uint64(b[i*8 : (i+1)*8])
	}
}

// LShift computes a left bit shift.
// From https://github.com/bitcoin/bitcoin/blob/master/src/arith_uint256.cpp#L21
func (u *uint256) LShift(n int) {
	tmp := *u

	k := n / 64
	shift := n % 64

	u.Zero()
	for i := 0; i < WIDTH; i++ {
		if i+k+1 < WIDTH && shift != 0 {
			u[i+k+1] |= (tmp[i] >> (64 - shift))
		}
		if i+k < WIDTH {
			u[i+k] |= (tmp[i] << shift)
		}
	}
}

// RShift computes a right bit shift.
// From https://github.com/bitcoin/bitcoin/blob/master/src/arith_uint256.cpp#L38
func (u *uint256) RShift(n int) {
	tmp := *u

	k := n / 64
	shift := n % 64

	u.Zero()
	for i := 0; i < WIDTH; i++ {
		if i-k-1 >= 0 && shift != 0 {
			u[i-k-1] |= (tmp[i] << (64 - shift))
		}
		if i-k >= 0 {
			u[i-k] |= (tmp[i] >> shift)
		}
	}
}

// Cmp compares two hash values.
// From https://github.com/bitcoin/bitcoin/blob/master/src/arith_uint256.cpp#L109
func (a *uint256) Cmp(b *uint256) int {
	for i := WIDTH - 1; i >= 0; i-- {
		switch {
		case a[i] < b[i]:
			return -1
		case a[i] > b[i]:
			return 1
		}
	}

	return 0
}

// CmpUint64 compares the current hash value with a 64-bit integer.
func (a *uint256) CmpUint64(b uint64) int {
	var b256 uint256
	b256.FromUint64(b)
	return a.Cmp(&b256)
}

// SetDifficulty sets the difficulty, by setting the target hash to 2^difficulty.
// Inspired by https://github.com/bitcoin/bitcoin/blob/master/src/arith_uint256.h#L277
// and https://github.com/bitcoin/bitcoin/blob/v0.1.5/bignum.h#L257, but severely
// simplified.
func (u *uint256) SetDifficulty(difficulty uint8) {
	u.FromUint64(1)
	u.LShift(int(difficulty))
}

// ToShortString converts the first 64 bits of a uint256 to a hex string.
func (u *uint256) ToShortString() string {
	return fmt.Sprintf("%016x", u[3])
}

// ToString converts a uint256 to a hex string.
func (u *uint256) ToString() (s string) {
	for _, n := range u {
		s = fmt.Sprintf("%016x", n) + s
	}
	return
}
