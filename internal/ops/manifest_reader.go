package ops

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func NewManifestReader(start <-chan bool, manifest, prefix string) (<-chan *EntryInfo, error) {

	freader, err := os.Open(manifest)
	if err != nil {
		return nil, err
	}

	out := make(chan *EntryInfo, 2)
	reader := manifest_reader{
		start:  start,
		out:    out,
		reader: freader,
		prefix: prefix,
	}
	go reader.run()

	return out, nil
}

type manifest_reader struct {
	start  <-chan bool
	out    chan<- *EntryInfo
	reader io.ReadCloser
	prefix string
}

func (mr *manifest_reader) run() {
	defer close(mr.out)
	defer mr.reader.Close()

	if v, ok := <-mr.start; !v || !ok {
		return
	}

	prefix := filepath.Clean(mr.prefix)

	scanner := bufio.NewScanner(mr.reader)
	for scanner.Scan() {

		// split the line into tokens
		line := scanner.Text()
		tokens := strings.Split(line, ",")
		if len(tokens) != 5 {
			continue
		}

		// convert the tokens to the correct type
		var size int64
		fmt.Sscanf(tokens[0], "%d", &size)

		var mtime int64
		fmt.Sscanf(tokens[1], "%d", &mtime)

		var mode os.FileMode
		fmt.Sscanf(tokens[2], "%o", &mode)

		hash := tokens[3]

		path, err := url.PathUnescape(tokens[4])
		if err != nil {
			continue
		}
		path = filepath.Join(prefix, filepath.Clean(path))

		// pass the message on
		ei := EntryInfo{
			Status:  StatusOk,
			Path:    path,
			Hash:    hash,
			Size:    size,
			ModTime: mtime,
			Mode:    mode,
		}
		mr.out <- &ei
	}
}
