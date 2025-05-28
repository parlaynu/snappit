package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/parlaynu/snappit/internal/arena"
	"github.com/parlaynu/snappit/internal/ops"
)

func RestoreSnapshot(config *Config, name string) error {

	arena, err := arena.New(config.Arena)
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
		fcount, err := restore_archive(snapshot, archive.Label, archive.Source, config.DeepScan, config.SkipDirs)
		if err != nil {
			return err
		}
		dcount := 0
		if config.Prune {
			dcount, err = prune_source(archive.Source, config.SkipDirs)
			if err != nil {
				return err
			}
		}
		fmt.Printf("checked %d files\n", fcount)
		if config.Prune {
			fmt.Printf("pruned %d directories\n", dcount)
		}
	}

	return nil
}

func restore_archive(snapshot arena.Snapshot, label, source string, deep_check bool, skipdirs []string,) (int, error) {
	fmt.Printf("Restoring %s:%s\n", label, source)

	archive, err := snapshot.LoadArchive(label)
	if err != nil {
		return 0, err
	}

	// create the start channel - an early return will
	//  close the channel and shut everything down
	start := make(chan bool, 2)
	defer close(start)

	// build the pipeline of operators
	ch1, err := ops.NewFsScanner(start, source, skipdirs)
	if err != nil {
		return 0, err
	}

	if deep_check {
		ch1, err = ops.NewHashGenerator(ch1)
		if err != nil {
			return 0, err
		}
	}

	manifest := archive.Manifest()
	ch2, err := ops.NewManifestReader(start, manifest, source)
	if err != nil {
		return 0, err
	}

	ch, err := ops.NewStreamMerger(ch1, ch2)
	if err != nil {
		return 0, err
	}
	ch, err = ops.NewFsReconciler(ch, source, archive.Data())
	if err != nil {
		return 0, err
	}

	// and start the pipeline
	start <- true
	start <- true

	count := 0
	for info := range ch {
		count++
		if info.Status != ops.StatusOk {
			fmt.Printf("- %s: %s\n", strings.ToLower(ops.EntryStatusName(info.Status)), info.Path)
		}
		if info.Status == ops.StatusError {
			fmt.Printf("  %v\n", info.Error)
		}
	}

	return count, nil
}

func prune_source(source string, skipdirs []string) (int, error) {

	// create the start channel - an early return will
	//  close the channel and shut everything down
	start := make(chan bool, 2)
	defer close(start)

	ch, err := ops.NewFsPruner(start, source, skipdirs)
	if err != nil {
		return 0, err
	}
	start <- true

	count := 0
	for info := range ch {
		if info.Pruned {
			count++
			fmt.Printf("- pruned: %s\n", info.Directory)
		}
	}

	return count, nil
}
