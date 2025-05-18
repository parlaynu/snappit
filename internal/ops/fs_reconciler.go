package ops

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func NewFsReconciler(in <-chan *EntryInfo, fs_root, data_root string) (<-chan *EntryInfo, error) {

	// make sure we have a trailing slash... assumed in the main loop
	if !strings.HasSuffix(fs_root, "/") {
		fs_root += "/"
	}

	out := make(chan *EntryInfo, 2)
	fs := fs_reconciler{
		in:        in,
		out:       out,
		fs_root:   fs_root,
		data_root: data_root,
	}
	go fs.run()

	return out, nil
}

type fs_reconciler struct {
	in        <-chan *EntryInfo
	out       chan<- *EntryInfo
	fs_root   string
	data_root string
}

func (fs *fs_reconciler) run() {
	defer close(fs.out)

	// loop over the source files
	for info := range fs.in {
		if info.Status == StatusModified || info.Status == StatusNotInFilesystem {
			err := fs.restore_file(info)
			if err != nil {
				info.Status = StatusError
				info.Error = err
			} else {
				info.Status = StatusRestored
			}
		} else if info.Status == StatusNotInManifest {
			err := fs.remove_file(info)
			if err != nil {
				info.Status = StatusError
				info.Error = err
			} else {
				info.Status = StatusRemoved
			}
		}
		fs.out <- info
	}
}

func (fs *fs_reconciler) restore_file(info *EntryInfo) error {
	// remove any existing file
	if info.Status == StatusModified {
		err := os.Remove(info.Path)
		if err != nil {
			return err
		}
	}

	// open the source
	dpath := fmt.Sprintf("%s/%s/%s", fs.data_root, info.Hash[:4], info.Hash)
	src, err := os.Open(dpath)
	if err != nil {
		return err
	}
	defer src.Close()

	// create the new destination file
	dst, err := os.Create(info.Path)
	if err != nil {
		return err
	}
	defer dst.Close()

	// and copy
	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	// reset the mode
	err = os.Chmod(info.Path, info.Mode)
	if err != nil {
		return err
	}

	// reset the file modify time
	err = os.Chtimes(info.Path, time.Now(), time.Unix(info.ModTime, 0))
	if err != nil {
		return err
	}

	return nil
}

func (fs *fs_reconciler) remove_file(info *EntryInfo) error {
	return os.Remove(info.Path)
}
