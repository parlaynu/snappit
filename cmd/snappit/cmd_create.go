package main

import (
	"fmt"

	"github.com/parlaynu/snappit/internal/arena"
	"github.com/parlaynu/snappit/internal/ops"
)

func CreateSnapshot(config *Config, name, baseline string) error {

	arena, err := arena.New(config.Arena)
	if err != nil {
		return err
	}
	err = arena.Create()
	if err != nil {
		return err
	}

	snapshot, err := arena.CreateSnapshot(name, baseline)
	if err != nil {
		return err
	}

	for _, archive := range config.Archives {
		err = create_archive(snapshot, archive.Label, archive.Source, config.SkipDirs)
		if err != nil {
			return err
		}
	}

	return nil
}

func create_archive(snapshot arena.Snapshot, label, source string, skipdirs []string) error {
	fmt.Printf("Archiving %s:%s\n", label, source)

	archive, err := snapshot.CreateArchive(label)
	if err != nil {
		return err
	}

	// create the start channel - an early return will
	//  close the channel and shut everything down
	start := make(chan bool, 2)
	defer close(start)

	// build the pipeline of operators
	ch, err := ops.NewFsScanner(start, source, skipdirs)
	if err != nil {
		return err
	}

	baseline := archive.Baseline()
	if len(baseline) > 0 {
		ch2, err := ops.NewManifestReader(start, baseline, source)
		if err != nil {
			return err
		}
		ch, err = ops.NewHashInserter(ch2, ch)
		if err != nil {
			return err
		}
	}

	ch, err = ops.NewHashGenerator(ch)
	if err != nil {
		return err
	}
	ch, err = ops.NewFileArhiver(ch, archive.Data())
	if err != nil {
		return err
	}
	ch, err = ops.NewManifestWriter(ch, archive.Manifest(), source)
	if err != nil {
		return err
	}

	// and start the pipeline
	start <- true
	if len(baseline) > 0 {
		start <- true
	}

	count := 0
	for v := range ch {
		count++
		if v.Status != ops.StatusOk {
			fmt.Printf("found %s: %s\n", v.Hash, v.Path)
		}
	}
	fmt.Printf("checked %d files\n", count)

	return nil
}
