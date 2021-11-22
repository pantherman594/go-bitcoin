package main

import (
  "bytes"
  "crypto/sha256"
  "encoding/binary"
)

type uint256 [sha256.Size]byte

func (u *uint256) FromUint64(o uint64) {
  var buf bytes.Buffer
  err := binary.Write(&buf, binary.BigEndian, o)
  if err != nil {
    panic(err)
  }

  copy(u[sha256.Size - 8:], buf.Bytes())
}

func (u *uint256) LShift(n int) {
  if n > 0 {
    copy(u[:], u[n:])
    copy(u[sha256.Size - n:], make([]byte, n))
  } else if n < 0 {
    copy(u[-n:], u[:sha256.Size + n])
    copy(u[:-n], make([]byte, -n))
  }
}

func (u *uint256) RShift(n int) {
  u.LShift(-n)
}

// Altered from https://cs.opensource.google/go/go/+/refs/tags/go1.17.3:src/math/big/nat.go;drc=refs%2Ftags%2Fgo1.17.3;l=166.
func (a *uint256) Cmp(b *uint256) (r int) {
  i := sha256.Size - 1
  for i > 0 && a[i] == b[i] {
    i--
  }

  switch {
  case a[i] < b[i]:
    r = -1
  case a[i] > b[i]:
    r = 1
  }

  return
}

func (a *uint256) CmpUint64(b uint64) int {
  var b256 uint256
  b256.FromUint64(b)
  return a.Cmp(&b256)
}

// arith_uint256.cp:203
func (u *uint256) SetCompact(compact uint32, negative *bool, overflow *bool) {
  // "number of bytes of N"
  size := compact >> 24

  // lower 23 bits are the mantissa
  word := compact & 0x007fffff

  if size <= 3 {
    word >>= 8 * (3 - size)
    u.FromUint64(uint64(word))
  } else {
    u.FromUint64(uint64(word))
    u.LShift(8 * (int(size)-3))
  }

  if *negative {
    *negative = word != 0 && (compact & 0x00800000) != 0
  }

  if *overflow {
    *overflow = word != 0 && (size > 34) ||
                (word > 0xff && size > 33) ||
                (word > 0xffff && size > 32)
  }

  return
}
