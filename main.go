package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

const (
	MIN_VERSION = 1
	VERSION     = 1
)

var (
	debug    = false
	faulty   = false
	maxTries = uint64(math.MaxUint64)
)

func dbgPrint(a ...interface{}) (n int, err error) {
	if debug {
		return fmt.Println(a...)
	}

	return 0, nil
}

// Creates a network of 1 logger and numMiners miners, and starts them.
func SetupNetwork(numMiners int, startingDifficulty uint8) error {
	// Set iniital min difficulty.
	var target uint256
	target.SetDifficulty(startingDifficulty)

	logger := Logger{}

	logger.blocks = make(map[uint256]Block)
	logger.minWorkRequired = target
	logger.recvBlock = make(chan Block)
	logger.sendBlocks = make([]chan Block, numMiners)

	fmt.Println("Starting miners...")
	for i := range logger.sendBlocks {
		logger.sendBlocks[i] = make(chan Block)
		miner := Miner{
			name:      int8(i),
			recvBlock: logger.sendBlocks[i],
			sendBlock: logger.recvBlock,
		}

		go miner.Start()
	}

	fmt.Println("Starting logger...")
	logger.Start(startingDifficulty)

	return nil
}

func main() {
	numMiners := flag.Int("n", 20,
		"The number of miners in the network.")
	difficulty := flag.Int("d", 233,
		"The difficulty, higher is easier, where the hash must be less than or equal to 2^difficulty. Max value is 255.")
	flag.BoolVar(&debug, "debug", false, "Enable extra debug printing.")
	flag.BoolVar(&faulty, "faulty", false, "Simulate faulty nodes.")

	flag.Parse()

	if *numMiners < 1 {
		log.Fatalf("Must have at least 1 miner.")
	}

	if *difficulty < 0 || *difficulty > 255 {
		log.Fatalf("Difficulty must be between 0 and 255 inclusive.")
	}

	rand.Seed(time.Now().UnixNano())

	SetupNetwork(*numMiners, uint8(*difficulty))
}
