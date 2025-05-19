package main

import (
	"fmt"
	"os"

	"github.com/parlaynu/snappit/internal/arena"
)

func ListSnapshots(config *Config) error {

	arena, err := arena.New(config.Arena)
	if err != nil {
		return err
	}
	if !arena.Exists() {
		return os.ErrNotExist
	}

	snaps, err := arena.Snapshots()
	if err != nil {
		return err
	}
	if len(snaps) == 0 {
		fmt.Printf("No snapshots found in %s\n", arena.Path())
		return nil
	}

	fmt.Println("Snapshots")
	for _, snap := range snaps {
		fmt.Printf("- %s\n", snap)
	}

	return nil
}
