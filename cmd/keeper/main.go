package main

import (
	"context"
	"flag"
	"log"

	"github.com/trigg3rX/go-backend/internal/keeper"
)

func main() {
	nameFlag := flag.String("name", "", "Name of the keeper (Frodo, Sam, Merry, or Pippin)")
	flag.Parse()

	if *nameFlag == "" {
		log.Fatal("Please provide a keeper name using -name flag")
	}

	ctx := context.Background()
	node, err := keeper.NewNode(ctx, *nameFlag)
	if err != nil {
		log.Fatal(err)
	}

	if err := node.Start(); err != nil {
		log.Fatal(err)
	}

	select {}
}