package main

import (
	"fmt"
	"runtime"
	"time"
)

func perf() {
	quiet = true

	maxMaxProcs := runtime.GOMAXPROCS(-1)

	numMiners := 1
	goMaxProcs := maxMaxProcs
	difficulty := uint8(240)
	numBlocks := 10


	/*
	maxMiners := 30
	fmt.Printf("Difficulty: %d, NumBlocks: %d, GoMaxProcs: %d\n", difficulty, numBlocks, goMaxProcs)
	fmt.Println("NumMiners\tTime (ns)")
	for ; numMiners < maxMiners; numMiners++ {
		totalTime, times := SetupNetwork(numMiners, difficulty, numBlocks, goMaxProcs, 5 * time.Minute)
		fmt.Printf("%d\t%d\n", numMiners, totalTime.Nanoseconds()/int64(len(times)))
	}

	fmt.Println()
	fmt.Println()

	numMiners = 6
	goMaxProcs = 1

	fmt.Printf("NumMiners: %d, Difficulty: %d, NumBlocks: %d\n", numMiners, difficulty, numBlocks)
	fmt.Println("GoMaxProcs\tTime (ns)")
	for ; goMaxProcs <= maxMaxProcs; goMaxProcs++ {
		totalTime, times := SetupNetwork(numMiners, difficulty, numBlocks, goMaxProcs, 5 * time.Minute)
		fmt.Printf("%d\t%d\n", goMaxProcs, totalTime.Nanoseconds()/int64(len(times)))
	}

	fmt.Println()
	fmt.Println()

	numMiners = 2
	goMaxProcs = 1

	fmt.Printf("NumMiners: %d, Difficulty: %d, NumBlocks: %d\n", numMiners, difficulty, numBlocks)
	fmt.Println("GoMaxProcs\tTime (ns)")
	for ; goMaxProcs <= maxMaxProcs; goMaxProcs++ {
		totalTime, times := SetupNetwork(numMiners, difficulty, numBlocks, goMaxProcs, 5 * time.Minute)
		fmt.Printf("%d\t%d\n", goMaxProcs, totalTime.Nanoseconds()/int64(len(times)))
	}
	*/

	fmt.Println()
	fmt.Println()

	numMiners = 12
	difficulty = 230
	goMaxProcs = 8

	fmt.Printf("NumMiners: %d, NumBlocks: %d, GoMaxProcs: %d\n", numMiners, numBlocks, goMaxProcs)
	fmt.Println("Difficulty\tTime (ns)")
	for ;; difficulty-- {
		totalTime, times := SetupNetwork(numMiners, difficulty, numBlocks, goMaxProcs, 5 * time.Minute)
		fmt.Printf("%d\t%d\n", difficulty, totalTime.Nanoseconds()/int64(len(times)))

		if len(times) < numBlocks {
			fmt.Println("Timed out")
			break
		}
	}
}
