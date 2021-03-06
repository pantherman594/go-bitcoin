package main

import (
	"fmt"
	"sync"
	"time"
)

type Logger struct {
	blocks    map[uint256]Block
	lastTime  time.Time
	bestChain uint64

	minWorkRequired uint256
	recvBlock       chan Block
	sendBlocks      []chan Block

	durations chan time.Duration
	minersClosed chan struct{}
	closed    chan struct{}
	minerWg   *sync.WaitGroup
	closeWg   *sync.WaitGroup
}

func (l *Logger) Start(startingDifficulty uint8) {
	defer l.closeWg.Done()

	if l.closed == nil {
		l.closed = make(chan struct{})
	}

	// Create an initial block.
	initialBlock := Block{
		BlockData: BlockData{
			version:    VERSION,
			data:       -1,
			difficulty: startingDifficulty,
		},
		valid:  true,
		height: 0,
	}
	l.blocks[initialBlock.hash()] = initialBlock

	// Send the new block to all miners.
	l.sendBlock(&initialBlock)
	l.lastTime = time.Now()

	// Process received solutions.
loop:
	for {
		select {
		case _, ok := <-l.closed:
			if !ok {
				break loop
			}
		case block, ok := <-l.recvBlock:
			if !ok {
				break loop
			}

			// Check and add the block to the chain.
			err := l.processBlock(&block)
			if err != nil {
				if !quiet {
					fmt.Println("Invalid block received:", err)
				}
			}
		}
	}

	close(l.minersClosed)
	done := make(chan struct{})
	go func(ch chan struct{}, wg *sync.WaitGroup) {
		wg.Wait()
		done <- struct{}{}
	}(done, l.minerWg)

	// Clear the recvBlock channel.
	for {
		select {
		case <-l.recvBlock:
		case <-done:
			// Finish when all miners are done
			return
		}
	}
}

// processBlock checks the solution, and adds it to the chain if it is valid.
func (l *Logger) processBlock(block *Block) error {
	// Check whether the block is correct.
	if block.version < MIN_VERSION || block.version > VERSION {
		return fmt.Errorf("Outdated version")
	}

	// Sanity check, only the logger should mark a block as valid.
	if block.valid {
		return fmt.Errorf("Recevied blocks should not be valid")
	}

	// Set the difficulty according to the block.
	var target uint256
	target.SetDifficulty(block.difficulty)

	// Make sure the difficulty is above the minimum difficulty.
	// Inspired by https://github.com/bitcoin/bitcoin/blob/v0.1.5/main.cpp#L1210
	// and https://github.com/bitcoin/bitcoin/blob/master/src/pow.cpp#L13, but
	// simplified to not require an increasing difficulty.
	if target.CmpUint64(0) == 0 || target.Cmp(&l.minWorkRequired) > 0 {
		return fmt.Errorf("Invalid target bits")
	}

	hash := block.hash()

	// Make sure the puzzle was actually solved.
	if hash.Cmp(&target) > 0 { // hash > target
		return fmt.Errorf("Invalid nonce, puzzle not solved")
	}

	// From https://github.com/bitcoin/bitcoin/blob/v0.1.5/main.cpp#L1242
	_, found := l.blocks[hash]
	if found {
		return fmt.Errorf("Block already found")
	}

	// Inspired by https://github.com/bitcoin/bitcoin/blob/v0.1.5/main.cpp#L1255,
	// but discarding orphan blocks.
	prev, found := l.blocks[block.hashPrevBlock]
	if !found {
		return fmt.Errorf("Invalid previous block")
	}

	// Set the block to valid and increase the chain height.
	block.valid = true
	block.height = prev.height + 1

	// Add the block to the chain.
	l.blocks[hash] = *block

	// Continue only if the chain height increases.
	if block.height <= l.bestChain {
		return fmt.Errorf("New block does not increase chain height")
	}

	l.bestChain = block.height
	timeNeeded := time.Since(l.lastTime)
	if l.durations != nil {
		select {
		case l.durations <- timeNeeded:
		default:
		}
	}

	if !quiet {
		// Print a status of the current longest chain.
		fmt.Print("\n\n\n")
		fmt.Println("======================")
		fmt.Println("Puzzle solved with hash", hash.ToString())
		fmt.Println("Time:", timeNeeded)
		fmt.Println("Difficulty:", target.ToShortString())
		fmt.Println()
		fmt.Print("Chain to current:")

		// Print out all the hashes in the chain.
		prev = *block
		found = true
		for found {
			hash := prev.hash()
			fmt.Print("\n", (&hash).ToShortString())
			prev, found = l.blocks[prev.hashPrevBlock]
		}
		fmt.Println(" (initial)")

		fmt.Println("======================")
		fmt.Print("\n\n\n")
	}

	l.lastTime = time.Now()

	// Send the block to all miners.
	l.sendBlock(block)

	return nil
}

func (l *Logger) sendBlock(block *Block) {
	for _, ch := range l.sendBlocks {
		ch <- *block
	}
}
