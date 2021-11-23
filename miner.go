package main

import (
	"fmt"
	"math/rand"
	"sync/atomic"
)

type Miner struct {
	name int8

	nonceOffset         uint64
	transactionsUpdated uint64

	recvBlock chan Block
	sendBlock chan Block
}

func (m *Miner) Start() {
	// Set a random offset.
	m.nonceOffset = rand.Uint64()

	// Process any received blocks.
	for block := range m.recvBlock {
		if !block.valid {
			dbgPrint("Received invalid block")
			continue
		}

		atomic.AddUint64(&m.transactionsUpdated, 1)

		// Create a new block, following the received block.
		newBlock := Block{
			BlockData: BlockData{
				version:       VERSION,
				hashPrevBlock: block.hash(),
				data:          m.name,
				difficulty:    block.difficulty,
			},
			valid: false,
		}

		if faulty && rand.Float32() < 0.01 {
			fmt.Println(m.name, "SEND REPEAT")
			block.valid = false
			m.sendBlock <- block
		}

		go m.mineBlock(&newBlock)
	}
}

// mineBlock mines a given block by attempting to hash it with different nonces,
// starting from a random offset.
func (m *Miner) mineBlock(block *Block) bool {
	// From https://github.com/bitcoin/bitcoin/blob/master/src/pow.cpp#L80
	var target uint256
	target.SetDifficulty(block.difficulty)

	if faulty && rand.Float32() < 0.01 {
		fmt.Println(m.name, "INVALID PREVIOUS")
		block.hashPrevBlock.FromUint64(rand.Uint64())
	}

	dbgPrint(m.name, "mining with prev", block.hashPrevBlock.ToShortString())
	if target.CmpUint64(0) == 0 {
		dbgPrint("Invalid target")
		return false
	}

	// Start testing at the miner's nonceOffset.
	block.nonce = m.nonceOffset

	// Store the current number of transactions.
	// From https://github.com/bitcoin/bitcoin/blob/v0.1.5/main.cpp#L2231
	transactionsUpdatedLast := m.transactionsUpdated

	remTries := maxTries

	// From https://github.com/bitcoin/bitcoin/blob/master/src/rpc/mining.cpp#L123
	// and https://github.com/bitcoin/bitcoin/blob/v0.1.5/main.cpp#L2339
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

		if faulty && (block.nonce&0x3ffff) == 0 {
			r := rand.Float32()
			if r < 0.005 {
				fmt.Println(m.name, "INCORRECT NONCE")
				m.sendBlock <- *block
			} else if r < 0.01 {
				fmt.Println(m.name, "INCORRECT DIFFICULTY")
				block.difficulty = 255
				target.SetDifficulty(255)
			} else if r < 0.015 {
				fmt.Println(m.name, "SEND VALID")
				block.valid = true
				m.sendBlock <- *block
			}
		}

		// Check for new transactions every few seconds.
		// From https://github.com/bitcoin/bitcoin/blob/v0.1.5/main.cpp#L2381
		if (block.nonce & 0x3ffff) == 0 {
			if atomic.LoadUint64(&m.transactionsUpdated) != transactionsUpdatedLast {
				dbgPrint(fmt.Sprintf("Miner %d aborting because new transaction received.", m.name))
				return false
			}
			dbgPrint(m.name, "mining... current nonce:", block.nonce)
		}
	}
	dbgPrint(m.name, "quit")

	return false
}
