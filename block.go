package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

type BlockData struct {
	version       int32
	hashPrevBlock uint256
	data          int8 // stores the miner that completed the block, with -1 being the logger.
	difficulty    uint8
	nonce         uint64
}

type Block struct {
	BlockData
	valid bool
	height uint64
}

// toBytes converts the Block struct into a slice of bytes.
func (b *BlockData) toBytes() []byte {
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
func (b *BlockData) hash() (u uint256) {
	hash1 := sha256.Sum256(b.toBytes())

	u.FromBytes(sha256.Sum256(hash1[:]))
	return
}
