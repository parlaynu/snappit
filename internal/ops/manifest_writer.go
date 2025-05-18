package ops

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func NewManifestWriter(in <-chan *EntryInfo, manifest, prefix string) (<-chan *EntryInfo, error) {

	if prefix[len(prefix)-1] != filepath.Separator {
		prefix += string(filepath.Separator)
	}
	prefix = filepath.ToSlash(prefix)

	writer, err := os.Create(manifest)
	if err != nil {
		return nil, err
	}

	out := make(chan *EntryInfo, 2)
	mw := manifest_writer{
		in:     in,
		out:    out,
		writer: writer,
		prefix: prefix,
	}
	go mw.run()

	return out, nil
}

type manifest_writer struct {
	in     <-chan *EntryInfo
	out    chan<- *EntryInfo
	writer io.WriteCloser
	prefix string
}

func (mw *manifest_writer) run() {
	defer close(mw.out)
	defer mw.writer.Close()

	for info := range mw.in {
		if info.Status != StatusError {
			line, err := mw.process(info)
			if err == nil && len(line) > 0 {
				_, err = io.WriteString(mw.writer, line)
			}
			if err != nil {
				info.Status = StatusError
				info.Error = err
			}
		}
		mw.out <- info
	}
}

func (mw *manifest_writer) process(info *EntryInfo) (string, error) {
	if info.Status == StatusError {
		return "", nil
	}

	// trim the prefix and encode
	path := filepath.ToSlash(info.Path)
	path = url.PathEscape(strings.TrimPrefix(path, mw.prefix))

	line := fmt.Sprintf("%d,%d,0%o,%s,%s\n",
		info.Size,
		info.ModTime,
		info.Mode&0777,
		info.Hash,
		path,
	)

	return line, nil
}
