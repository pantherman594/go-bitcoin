package main

import (
	"fmt"
)

type Logger struct {
	blocks map[uint256]Block

	minWorkRequired uint256

	recvBlock  chan Block
	sendBlocks []chan Block
	bestChain uint64
}

func (l *Logger) Start() {
	// Create an initial block.
	initialBlock := Block{
		BlockData: BlockData{
			version:    VERSION,
			data:       -1,
			difficulty: STARTING_DIFFICULTY,
		},
		valid: true,
		height: 0,
	}
	l.blocks[initialBlock.hash()] = initialBlock

	// Send the new block to all miners.
	l.sendBlock(&initialBlock)

	for block := range l.recvBlock {
		// Check and add the block to the chain.
		err := l.processBlock(&block)
		if err != nil {
			fmt.Println("Invalid block received:", err)
			continue
		}
	}
}

func (l *Logger) processBlock(block *Block) error {
	// Check whether the block is correct.
	if block.version < MIN_VERSION {
		return fmt.Errorf("Outdated version")
	}

	if block.valid {
		return fmt.Errorf("Recevied blocks should not be valid")
	}

	var target uint256
	target.SetDifficulty(block.difficulty)

	if target.CmpUint64(0) == 0 || target.Cmp(&l.minWorkRequired) < 0 {
		return fmt.Errorf("Invalid target bits")
	}

	hash := block.hash()

	if hash.Cmp(&target) > 0 { // hash > target
		return fmt.Errorf("Invalid nonce, puzzle not solved")
	}

	_, found := l.blocks[hash]
	if found {
		return fmt.Errorf("Block already found")
	}

	prev, found := l.blocks[block.hashPrevBlock]
	if !found {
		return fmt.Errorf("Invalid previous block")
	}

	// Set the block to valid and increase the chain height.
	block.valid = true
	block.height = prev.height + 1

	if block.height <= l.bestChain {
		return fmt.Errorf("New block does not increase chain height")
	}

	l.bestChain = block.height

	// Add the block to the chain.
	l.blocks[hash] = *block

	fmt.Print("\n\n\n")
	fmt.Println("======================")
	fmt.Println("Puzzle solved with hash", hash.ToString())
	fmt.Println("Difficulty:", target.ToShortString())
	fmt.Println()
	fmt.Println("Chain to current:")

	prev = *block
	found = true
	for found {
		hash := prev.hash()
		fmt.Println((&hash).ToShortString())
		prev, found = l.blocks[prev.hashPrevBlock]
	}
	fmt.Println("(ROOT)")

	fmt.Println("======================")
	fmt.Print("\n\n\n")

	// Send the block to all miners.
	l.sendBlock(block)

	return nil
}

func (l *Logger) sendBlock(block *Block) {
	for _, ch := range l.sendBlocks {
		ch <- *block
	}
}
