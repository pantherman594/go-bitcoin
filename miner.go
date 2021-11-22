package main

import (
  "math/rand"
  "sync/atomic"
  "time"
)

type Miner struct {
  nonceOffset uint64
  maxTries uint64
  transactionsUpdated uint64

  recvBlock chan Block
  sendBlock chan Block
}

func (m *Miner) Start() {
  // Set a random offset.
  rand.Seed(time.Now().UnixNano())
  m.nonceOffset = rand.Uint64()

  // Process any received blocks.
  for block := range m.recvBlock {
    atomic.AddUint64(&m.transactionsUpdated, 1)
    go m.mineBlock(&block)
  }
}

// also from pow.cpp:74
func (m *Miner) mineBlock(block *Block) bool {
  var negative bool
  var overflow bool
  var target uint256
  target.SetCompact(block.bits, &negative, &overflow)

  if negative || target.CmpUint64(0) == 0 || overflow {
    return false
  }

  // Start testing at the miner's nonceOffset.
  block.nonce = m.nonceOffset

  
  transactionsUpdatedLast := m.transactionsUpdated

  remTries := m.maxTries

  for remTries > 0 {
    hash := block.hash()

    if hash.Cmp(&target) < 1 { // hash <= target
      // A nonce was found.

      block.nonce = block.nonce
      m.sendBlock <- *block
      return true
    }

    block.nonce += 1
    remTries -= 1

    // Check for new transactions every few seconds.
    if (block.nonce & 0x3ffff) == 0 {

      if atomic.LoadUint64(&m.transactionsUpdated) != transactionsUpdatedLast {
        return false
      }
    }
  }
  
  return false
}
