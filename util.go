package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const WIDTH = 4

type uint256 [WIDTH]uint64

func (u *uint256) Zero() {
	for i := 0; i < WIDTH; i++ {
		u[i] = 0
	}
}

func (u *uint256) FromUint64(o uint64) {
	u.Zero()
	u[0] = o
}

func (u *uint256) FromBytes(b [sha256.Size]byte) {
	u.Zero()
	for i := 0; i < WIDTH; i++ {
		u[i] = binary.LittleEndian.Uint64(b[i*8 : (i+1)*8])
	}
}

// From arith_uint256.cpp:20
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

// From arith_uint256.cpp:37
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

// From arith_uint256.cpp:108
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

func (a *uint256) CmpUint64(b uint64) int {
	var b256 uint256
	b256.FromUint64(b)
	return a.Cmp(&b256)
}

func (u *uint256) SetDifficulty(difficulty uint8) {
	u.FromUint64(1)
	u.LShift(int(difficulty))
}

func (u *uint256) ToShortString() string {
	return fmt.Sprintf("%016x", u[3])
}

func (u *uint256) ToString() (s string) {
	for _, n := range u {
		s = fmt.Sprintf("%016x", n) + s
	}
	return
}
