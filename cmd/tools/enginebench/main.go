package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"xiangqi-lab/internal/domain/xiangqi"
	"xiangqi-lab/internal/engine"
	"xiangqi-lab/internal/engine/builtin"
	"xiangqi-lab/internal/engine/difficulty"
)

func main() {
	level := flag.Int("level", 5, "difficulty level 1-10")
	fen := flag.String("fen", xiangqi.InitialFEN, "position FEN")
	flag.Parse()
	position, err := xiangqi.ParseFEN(*fen)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	profile := difficulty.Get(*level)
	result, err := builtin.New().Analyze(context.Background(), engine.AnalyzeRequest{
		Position: position, MaxDepth: profile.MaxDepth, MaxNodes: profile.MaxNodes,
		MoveTime: profile.MoveTime, MultiPV: profile.MultiPV,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"profile": profile, "result": result})
}
