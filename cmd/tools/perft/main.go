package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"xiangqi-lab/internal/domain/xiangqi"
)

func main() {
	fen := flag.String("fen", xiangqi.InitialFEN, "position FEN")
	depth := flag.Int("depth", 2, "search depth")
	flag.Parse()
	if *depth < 0 || *depth > 8 {
		fmt.Fprintln(os.Stderr, "depth must be between 0 and 8")
		os.Exit(2)
	}
	position, err := xiangqi.ParseFEN(*fen)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	started := time.Now()
	nodes := perft(position, *depth)
	fmt.Printf("depth=%d nodes=%d duration=%s\n", *depth, nodes, time.Since(started))
}

func perft(position xiangqi.Position, depth int) uint64 {
	if depth == 0 {
		return 1
	}
	var nodes uint64
	for _, move := range position.LegalMoves() {
		next, _, err := position.Apply(move)
		if err == nil {
			nodes += perft(next, depth-1)
		}
	}
	return nodes
}
