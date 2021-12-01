package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
)

type Miner struct {
	name int8

	nonceOffset         uint64
	transactionsUpdated uint64

	recvBlock chan Block
	sendBlock chan Block

	closed      chan struct{}
	activeMines sync.WaitGroup
	minerWg     *sync.WaitGroup
}

func (m *Miner) Start() {
	defer func() {
		m.activeMines.Wait()
		m.minerWg.Done()
	}()

	if m.closed == nil {
		m.closed = make(chan struct{})
	}

	// Process any received blocks.
loop:
	for {
		select {
		case _, ok := <-m.closed:
			if !ok {
				atomic.StoreUint64(&m.transactionsUpdated, math.MaxUint64)
				break loop
			}
		case block, ok := <-m.recvBlock:
			if !ok {
				break loop
			}

			if !block.valid {
				if !quiet {
					dbgPrint("Received invalid block")
				}
				continue loop
			}

			last := atomic.AddUint64(&m.transactionsUpdated, 1)

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
				if !quiet {
					fmt.Println(m.name, "SEND REPEAT")
				}
				block.valid = false
				m.sendBlock <- block
			}

			m.activeMines.Add(1)
			go m.mineBlock(last, &newBlock)
		}
	}
}

// mineBlock mines a given block by attempting to hash it with different nonces,
// starting from a random offset.
// Last stores the current number of transactions, from
// https://github.com/bitcoin/bitcoin/blob/v0.1.5/main.cpp#L2231
func (m *Miner) mineBlock(last uint64, block *Block) bool {
	defer m.activeMines.Done()

	// From https://github.com/bitcoin/bitcoin/blob/master/src/pow.cpp#L80
	var target uint256
	target.SetDifficulty(block.difficulty)

	if faulty && rand.Float32() < 0.01 {
		if !quiet {
			fmt.Println(m.name, "INVALID PREVIOUS")
		}
		block.hashPrevBlock.FromUint64(rand.Uint64())
	}

	dbgPrint(m.name, "mining with prev", block.hashPrevBlock.ToShortString())
	if target.CmpUint64(0) == 0 {
		dbgPrint("Invalid target")
		return false
	}

	// Start testing at the miner's nonceOffset.
	block.nonce = m.nonceOffset

	remTries := maxTries

	// From https://github.com/bitcoin/bitcoin/blob/master/src/rpc/mining.cpp#L123
	// and https://github.com/bitcoin/bitcoin/blob/v0.1.5/main.cpp#L2339
	for remTries > 0 {
		hash := block.hash()

		if hash.Cmp(&target) < 1 { // hash <= target
			// A nonce was found.
			if !quiet {
				fmt.Printf("Miner %d found a block with nonce %d and %d transactions.\n", m.name, block.nonce, last)
			}

			return m.sendIfNotClosed(last, block)
		}

		block.nonce += 1
		remTries -= 1

		if faulty && (block.nonce&0x3ffff) == 0 {
			r := rand.Float32()

			if r < 0.005 {
				if !quiet {
					fmt.Println(m.name, "INCORRECT NONCE")
				}

				if !m.sendIfNotClosed(last, block) {
					return false
				}
			} else if r < 0.01 {
				if !quiet {
					fmt.Println(m.name, "INCORRECT DIFFICULTY")
				}
				block.difficulty = 255
				target.SetDifficulty(255)
			} else if r < 0.015 {
				block.valid = true

				if !quiet {
					fmt.Println(m.name, "SEND VALID")
				}

				if !m.sendIfNotClosed(last, block) {
					return false
				}
			}
		}

		// Check for new transactions every few seconds.
		// From https://github.com/bitcoin/bitcoin/blob/v0.1.5/main.cpp#L2381
		if (block.nonce & 0x3ffff) == 0 {
			if atomic.LoadUint64(&m.transactionsUpdated) != last {
				dbgPrint(fmt.Sprintf("Miner %d aborting because new transaction received.", m.name))
				return false
			}
			dbgPrint(m.name, "mining... current nonce:", block.nonce)
		}
	}
	dbgPrint(m.name, "quit")

	return false
}

func (m *Miner) sendIfNotClosed(last uint64, block *Block) (r bool) {
	r = atomic.LoadUint64(&m.transactionsUpdated) == last
	if !r {
		dbgPrint(fmt.Sprintf("Miner %d aborting because new transaction received.", m.name))
	} else {
		m.sendBlock <- *block
	}

	return
}
