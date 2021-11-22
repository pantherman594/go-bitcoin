package main

import (
	"fmt"
	"math/rand"
	"sync/atomic"
)

type Miner struct {
	name int8

	nonceOffset         uint64
	maxTries            uint64
	transactionsUpdated uint64

	recvBlock chan Block
	sendBlock chan Block
}

func (m *Miner) Start() {
	// Set a random offset.
	m.nonceOffset = rand.Uint64()

	// Process any received blocks.
	for blocktx := range m.recvBlock {
		if !blocktx.valid {
			fmt.Println("Received invalid block")
			continue
		}

		atomic.AddUint64(&m.transactionsUpdated, 1)

		// Create a new block, following the received block.
		newBlock := Block{
			BlockData: BlockData{
				version:       VERSION,
				hashPrevBlock: blocktx.hash(),
				data:          m.name,
				difficulty:    blocktx.difficulty,
			},
			valid: false,
		}

		go m.mineBlock(&newBlock)
	}
}

// also from pow.cpp:74
func (m *Miner) mineBlock(block *Block) bool {
	var target uint256
	target.SetDifficulty(block.difficulty)

	fmt.Println(m.name, "mining with prev", block.hashPrevBlock.ToShortString())
	if target.CmpUint64(0) == 0 {
		fmt.Println("Invalid target")
		return false
	}

	// Start testing at the miner's nonceOffset.
	block.nonce = m.nonceOffset

	// Store the current number of transactions.
	transactionsUpdatedLast := m.transactionsUpdated

	remTries := m.maxTries

	for remTries > 0 {
		hash := block.hash()

		if hash.Cmp(&target) < 1 { // hash <= target
			// A nonce was found.
			fmt.Printf("Miner %d found a block with nonce %d.\n", m.name, block.nonce)
			m.sendBlock <- *block
			return true
		}

		block.nonce += 1
		remTries -= 1

		// Check for new transactions every few seconds.
		if (block.nonce & 0x3ffff) == 0 {
			if atomic.LoadUint64(&m.transactionsUpdated) != transactionsUpdatedLast {
				fmt.Printf("Miner %d aborting because new transaction received.\n", m.name)
				return false
			}
			fmt.Println(m.name, "mining... current nonce:", block.nonce)
		}
	}
	fmt.Println(m.name, "quit")

	return false
}
