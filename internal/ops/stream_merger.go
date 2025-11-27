package ops

import (
	"os"
	"strings"
)

func NewStreamMerger(inFsys <-chan *EntryInfo, inMani <-chan *EntryInfo) (<-chan *EntryInfo, error) {

	out := make(chan *EntryInfo, 2)
	sm := stream_merger{
		inFsys: inFsys,
		inMani: inMani,
		out:    out,
	}
	go sm.run()

	return out, nil
}

type stream_merger struct {
	inFsys <-chan *EntryInfo
	inMani <-chan *EntryInfo
	out    chan<- *EntryInfo
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func compare_paths(path1, path2 string) int {
	path1s := strings.Split(path1, string(os.PathSeparator))
	path2s := strings.Split(path2, string(os.PathSeparator))

	len1 := len(path1s)
	len2 := len(path2s)

	minlen := min(len1, len2)

	for i := 0; i < minlen; i++ {
		if path1s[i] < path2s[i] {
			return -1
		}
		if path1s[i] > path2s[i] {
			return 1
		}
	}

	return 0
}

func (sm *stream_merger) run() {
	defer close(sm.out)

	var fsysInfo *EntryInfo
	var maniInfo *EntryInfo

	for {
		// read from the fsys channel
		if fsysInfo == nil {
			if info, ok := <-sm.inFsys; ok {
				fsysInfo = info
			}
		}

		// read from the manifest channel
		if maniInfo == nil {
			if info, ok := <-sm.inMani; ok {
				maniInfo = info
			}
		}

		// both filesystem and manifest scans are finished: all done
		if fsysInfo == nil && maniInfo == nil {
			break
		}

		// manifest scan completed but not the filesystem: a new item
		if fsysInfo != nil && maniInfo == nil {
			fsysInfo.Status = StatusNotInManifest
			sm.out <- fsysInfo
			fsysInfo = nil
			continue
		}

		// filesystem scan completed but not the manifest: a removed item
		if fsysInfo == nil && maniInfo != nil {
			maniInfo.Status = StatusNotInFilesystem
			sm.out <- maniInfo
			maniInfo = nil
			continue
		}

		// compare the paths by path segment, not as strings
		val := compare_paths(fsysInfo.Path, maniInfo.Path)

		// if fsysInfo is behind maniInfo: a new item
		// - fsysInfo.Path < maniInfo.Path
		if val < 0 {
			fsysInfo.Status = StatusNotInManifest
			sm.out <- fsysInfo
			fsysInfo = nil
			continue
		}

		// if fsysInfo is ahead of maniInfo: a removed item
		// - fsysInfo.Path > maniInfo.Path
		if val > 0 {
			maniInfo.Status = StatusNotInFilesystem
			sm.out <- maniInfo
			maniInfo = nil
			continue
		}

		// if the paths are the same, check the attributes
		// - fsysInfo.Path == maniInfo.Path
		if val == 0 {
			maniInfo.Status = StatusOk

			// if size or modtime are different, flag as changed (or potentially changed)
			if fsysInfo.Size != maniInfo.Size || fsysInfo.ModTime != maniInfo.ModTime {
				maniInfo.Status = StatusModified
			}

			// if we have a filesystem hash, compare it to the manifest hash, potentially
			//   overriding the previous check's decision
			if len(fsysInfo.Hash) > 0 {
				maniInfo.Status = StatusOk
				if fsysInfo.Hash != maniInfo.Hash {
					maniInfo.Status = StatusModified
				}
			}
			sm.out <- maniInfo
			fsysInfo = nil
			maniInfo = nil
		}
	}
}
