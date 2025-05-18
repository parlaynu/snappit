package main

import (
	"fmt"
	"os"

	"github.com/parlaynu/snappit/internal/arena"
	"github.com/parlaynu/snappit/internal/ops"
)

func ResetSnapshot(config *Config, name string, verbose bool) error {

	arena, err := arena.New(config.Arena)
	if err != nil {
		return err
	}
	err = arena.Create()
	if err != nil {
		return err
	}
	if !arena.Exists() {
		return os.ErrNotExist
	}

	snapshot, err := arena.LoadSnapshot(name)
	if err != nil {
		return err
	}

	for _, archive := range config.Archives {
		err = reset_archive(snapshot, archive.Label, archive.Source, verbose)
		if err != nil {
			return err
		}
	}

	return nil
}

func reset_archive(snapshot arena.Snapshot, label, source string, verbose bool) error {
	fmt.Printf("Resetting %s:%s\n", label, source)

	archive, err := snapshot.LoadArchive(label)
	if err != nil {
		return err
	}

	// create the start channel - an early return will
	//  close the channel and shut everything down
	start := make(chan bool, 2)
	defer close(start)

	// build the pipeline of operators
	ch1, err := ops.NewFsScanner(start, source)
	if err != nil {
		return err
	}
	ch1, err = ops.NewHashGenerator(ch1)
	if err != nil {
		return err
	}

	manifest := archive.Manifest()
	ch2, err := ops.NewManifestReader(start, manifest, source)
	if err != nil {
		return err
	}

	ch, err := ops.NewStreamMerger(ch1, ch2)
	if err != nil {
		return err
	}
	ch, err = ops.NewFsReconciler(ch, source, archive.Data())
	if err != nil {
		return err
	}

	// and start the pipeline
	start <- true
	start <- true

	count := 0
	for info := range ch {
		count++
		if info.Status != ops.StatusOk {
			fmt.Printf("%s: %s\n", ops.EntryStatusName(info.Status), info.Path)
		}
	}
	fmt.Printf("checked %d files\n", count)

	return nil
}
