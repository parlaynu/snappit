package ops

import (
	"os"
	"path/filepath"
	"strings"
)

func NewFsScanner(start <-chan bool, source string, skipdirs []string) (<-chan *EntryInfo, error) {

	// make sure we have a trailing slash... assumed in the main loop
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}

	// create the skipdirs map
	skipdirz := make(map[string]bool)
	for _, dir := range skipdirs {
		skipdirz[dir] = true
	}

	out := make(chan *EntryInfo, 2)
	fs := fs_scanner{
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

type fs_scanner struct {
	out      chan<- *EntryInfo
	source   string
	skipdirs map[string]bool
}

func (fs *fs_scanner) run(dir string) {
	// read the directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// loop over the source files
	for _, entry := range entries {

		if entry.Type().IsRegular() {

			fpath := filepath.Join(dir, entry.Name())

			info, err := entry.Info()
			if err != nil {
				continue
			}

			fs.out <- &EntryInfo{
				Status:  StatusNew,
				Path:    fpath,
				Size:    info.Size(),
				ModTime: info.ModTime().Unix(),
				Mode:    info.Mode(),
			}

		} else if entry.Type().IsDir() {
			if !fs.skipdirs[entry.Name()] {
				fpath := filepath.Join(dir, entry.Name())
				fs.run(fpath)
			}
		}
	}
}
