package ops

func NewHashInserter(baseline, in <-chan *EntryInfo) (<-chan *EntryInfo, error) {
	out := make(chan *EntryInfo, 2)
	hi := hash_inserter{
		baseline: baseline,
		in:       in,
		out:      out,
	}
	go hi.run()

	return out, nil
}

type hash_inserter struct {
	baseline <-chan *EntryInfo
	in       <-chan *EntryInfo
	out      chan<- *EntryInfo
}

func (hi *hash_inserter) run() {
	defer close(hi.out)

	// read all the reference into a map
	ref := make(map[string]*EntryInfo)
	for info := range hi.baseline {
		ref[info.Path] = info
	}

	// if the ref map has the same path and the file size and modify time
	//   are the same, assume it's the same file and insert the hash
	for info := range hi.in {
		if info.Status != StatusError {
			ref_info, ok := ref[info.Path]
			if ok && ref_info.Size == info.Size && ref_info.ModTime == info.ModTime {
				info.Status = StatusOk
				info.Hash = ref_info.Hash
			}
		}
		hi.out <- info
	}
}
