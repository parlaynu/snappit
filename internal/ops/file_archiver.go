package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func NewFileArhiver(in <-chan *EntryInfo, root string) (<-chan *EntryInfo, error) {
	out := make(chan *EntryInfo, 2)
	fa := file_archiver{
		in:   in,
		out:  out,
		root: root,
	}
	go fa.run()

	return out, nil
}

type file_archiver struct {
	in   <-chan *EntryInfo
	out  chan<- *EntryInfo
	root string
}

func (fa *file_archiver) run() {
	defer close(fa.out)

	for info := range fa.in {
		if info.Status != StatusError {
			info, err := fa.process(info)
			if err != nil {
				info.Status = StatusError
				info.Error = err
			}
		}
		fa.out <- info
	}
}

func (fa *file_archiver) process(info *EntryInfo) (*EntryInfo, error) {

	// generate the storage path
	dpath := fmt.Sprintf("%s/%s/%s", fa.root, info.Hash[:4], info.Hash)

	// if it already exists, we can finish
	fi, err := os.Stat(dpath)
	if err != nil && !os.IsNotExist(err) {
		return info, err
	}
	if err == nil && fi.Size() == info.Size {
		info.Status = StatusOk
		return info, nil
	}

	// open source
	src, err := os.Open(info.Path)
	if err != nil {
		return info, err
	}
	defer src.Close()

	// create the destination
	dst, err := os.Create(dpath)
	if err != nil {
		os.MkdirAll(filepath.Dir(dpath), 0750)
		dst, err = os.Create(dpath)
	}
	if err != nil {
		return info, err
	}
	defer dst.Close()

	// copy the file
	_, err = io.Copy(dst, src)
	if err != nil {
		return info, err
	}

	return info, nil
}
