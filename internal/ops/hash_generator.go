package ops

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func NewHashGenerator(in <-chan *EntryInfo) (<-chan *EntryInfo, error) {
	out := make(chan *EntryInfo, 2)
	hg := hash_generator{
		in:  in,
		out: out,
	}
	go hg.run()

	return out, nil
}

type hash_generator struct {
	in  <-chan *EntryInfo
	out chan<- *EntryInfo
}

func (hg *hash_generator) run() {
	defer close(hg.out)

	for info := range hg.in {
		if info.Status != StatusError {
			info, err := hg.process(info)
			if err != nil {
				info.Status = StatusError
				info.Error = err
			}
		}
		hg.out <- info
	}
}

func (hg *hash_generator) process(info *EntryInfo) (*EntryInfo, error) {

	// only re-generate the hash if we need to
	if len(info.Hash) != 0 {
		return info, nil
	}

	// open for reading
	in, err := os.Open(info.Path)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	// generate the hash
	h := sha256.New()
	_, err = io.Copy(h, in)
	if err != nil {
		return nil, err
	} else {
		info.Hash = hex.EncodeToString(h.Sum(nil))
	}

	return info, nil
}
