package main

import (
  "bytes"
  "crypto/sha256"
  "encoding/binary"
)

type Block struct {
  version int32
  hashPrevBlock uint256
  hashTransaction uint256
  time uint32
  bits uint32
  nonce uint64
}

// toBytes converts the Block struct into a slice of bytes.
func (b *Block) toBytes() []byte {
  var buf bytes.Buffer
  err := binary.Write(&buf, binary.LittleEndian, b)
  if err != nil {
    panic(err)
  }

  return buf.Bytes()
}

// hash double-sha256 hashes the data in the Block. Double hashing, as also
// done in the original bitcoin code (main.cpp:2341), prevents length extension
// attacks.
func (b *Block) hash() uint256 {
  hash1 := sha256.Sum256(b.toBytes())
  return sha256.Sum256(hash1[:])
}

