package ops

import (
	"os"
	"path/filepath"
)

type PruneInfo struct {
	Directory string
	Pruned    bool
}

func NewFsPruner(start <-chan bool, source string, skipdirs []string) (<-chan *PruneInfo, error) {

	// create the skipdirs map
	skipdirz := make(map[string]bool)
	for _, dir := range skipdirs {
		skipdirz[dir] = true
	}

	out := make(chan *PruneInfo, 2)
	fs := fs_pruner{
		out:      out,
		source:   source,
		skipdirs: skipdirz,
	}
	go func() {
		defer close(fs.out)
		if v, ok := <-start; !v || !ok {
			return
		}
		fs.run(fs.source)
	}()

	return out, nil
}

type fs_pruner struct {
	out      chan<- *PruneInfo
	source   string
	skipdirs map[string]bool
}

func (fs *fs_pruner) run(dir string) int {
	// read the directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return -1
	}

	// loop over the source files
	npruned := 0
	for _, entry := range entries {
		if entry.Type().IsDir() {
			if fs.skipdirs[entry.Name()] {
				continue
			}
			fpath := filepath.Join(dir, entry.Name())

			// traverse into the subdirectory
			prune := fs.run(fpath) == 0
			if prune && os.Remove(fpath) == nil {
				npruned++
			}
			fs.out <- &PruneInfo{
				Directory: fpath,
				Pruned:    prune,
			}
		}
	}

	return len(entries) - npruned
}
