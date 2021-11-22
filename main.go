package main

import (
	"math"
	"math/rand"
	"time"
)

const (
	STARTING_DIFFICULTY = 233
	MIN_VERSION         = 1
	VERSION             = 1
	MAX_TRIES           = math.MaxUint64
)

func SetupNetwork(numMiners int) error {
	// Set iniital min difficulty, copying Satoshi's genesis block difficulty
	var target uint256
	target.SetDifficulty(STARTING_DIFFICULTY)

	logger := Logger{}

	logger.blocks = make(map[uint256]Block)
	logger.minWorkRequired = target
	logger.recvBlock = make(chan Block)
	logger.sendBlocks = make([]chan Block, numMiners)

	for i := range logger.sendBlocks {
		logger.sendBlocks[i] = make(chan Block)
		miner := Miner{
			name:      int8(i),
			maxTries:  MAX_TRIES,
			recvBlock: logger.sendBlocks[i],
			sendBlock: logger.recvBlock,
		}

		go miner.Start()
	}

	logger.Start()

	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	SetupNetwork(20)
}
