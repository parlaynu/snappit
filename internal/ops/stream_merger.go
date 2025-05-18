package ops

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

		// if fsysInfo is behind maniInfo
		if fsysInfo.Path < maniInfo.Path {
			fsysInfo.Status = StatusNotInManifest
			sm.out <- fsysInfo
			fsysInfo = nil
			continue
		}

		// if fsysInfo is ahead of maniInfo: a removed item
		if fsysInfo.Path > maniInfo.Path {
			maniInfo.Status = StatusNotInFilesystem
			sm.out <- maniInfo
			maniInfo = nil
			continue
		}

		// if the paths are the same, check the attributes
		if fsysInfo.Path == maniInfo.Path {
			// initialise the status and hash
			maniInfo.Status = StatusOk
			maniInfo.Hash = maniInfo.Hash

			// if size or modtime are different, flag as changed (or potentially changed)
			if fsysInfo.Size != maniInfo.Size || fsysInfo.ModTime != maniInfo.ModTime {
				maniInfo.Status = StatusModified
			}
			sm.out <- maniInfo
			fsysInfo = nil
			maniInfo = nil
		}
	}
}
