package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

const (
	MIN_VERSION = 1
	VERSION     = 1
)

var (
	debug    = false
	quiet    = false
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
func SetupNetwork(numMiners int, startingDifficulty uint8, limit int, maxProcs int, timeout time.Duration) (time.Duration, []time.Duration) {
	runtime.GOMAXPROCS(maxProcs)
	// Set iniital min difficulty.
	var target uint256
	target.SetDifficulty(startingDifficulty)

	closed := make(chan struct{})
	minersClosed := make(chan struct{})
	closeWg := sync.WaitGroup{}
	closeWg.Add(1)
	minerWg := sync.WaitGroup{}
	minerWg.Add(numMiners)

	logger := Logger{}

	logger.blocks = make(map[uint256]Block)
	logger.minWorkRequired = target
	logger.recvBlock = make(chan Block)
	logger.sendBlocks = make([]chan Block, numMiners)
	logger.closed = closed
	logger.minersClosed = minersClosed
	logger.closeWg = &closeWg
	logger.minerWg = &minerWg
	logger.durations = make(chan time.Duration)

	if !quiet {
		fmt.Println("Starting miners...")
	}
	for i := range logger.sendBlocks {
		logger.sendBlocks[i] = make(chan Block, 1)
		miner := Miner{
			name:      int8(i),
			recvBlock: logger.sendBlocks[i],
			sendBlock: logger.recvBlock,
			closed:    minersClosed,
			minerWg:   &minerWg,
			nonceOffset: uint64(i) * (math.MaxUint64 / uint64(numMiners)),
		}

		go miner.Start()
	}

	if !quiet {
		fmt.Println("Starting logger...")
	}
	go logger.Start(startingDifficulty)

	times := make([]time.Duration, 0)
	var totalTime time.Duration
	var i int64

loop:
	for i = 0; limit < 1 || i < int64(limit); i++ {
		select {
		case dur := <-logger.durations:
			totalTime += dur
			times = append(times, dur)
		case <-time.After(timeout):
			break loop
		}
	}

	close(closed)
	closeWg.Wait()

	for _, ch := range logger.sendBlocks {
		close(ch)
	}

	close(logger.recvBlock)

	return totalTime, times
}

func main() {
	numMiners := flag.Int("n", 20,
		"The number of miners in the network.")
	difficulty := flag.Int("d", 233,
		"The difficulty, higher is easier, where the hash must be less than or equal to 2^difficulty. Max value is 255.")
	limit := flag.Int("l", -1,
		"The number of blocks to mine. If limit < 1, the program will continue until stopped with Ctrl+C.")
	flag.BoolVar(&faulty, "faulty", false,
		"Simulates faulty miners (sending repeat solutions, setting an invalid previous hash, sending an unsolved puzzle, setting an incorrect difficulty, and sending a falsely \"valid\" block).")
	flag.BoolVar(&debug, "debug", false, "Enable extra debug printing.")
	flag.BoolVar(&quiet, "quiet", false, "Disable printing.")

	flag.Parse()

	if flag.NArg() > 0 && flag.Args()[0] == "perf" {
		perf()
		return
	}

	if quiet {
		debug = false
	}

	if *numMiners < 1 {
		log.Fatalf("Must have at least 1 miner.")
	}

	if *difficulty < 0 || *difficulty > 255 {
		log.Fatalf("Difficulty must be between 0 and 255 inclusive.")
	}

	rand.Seed(time.Now().UnixNano())

	SetupNetwork(*numMiners, uint8(*difficulty), *limit, -1, 5 * time.Minute)
}
