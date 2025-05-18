package arena

import (
	"os"
	"path/filepath"
)

type Snapshot interface {
	CreateArchive(label string) (Archive, error)
	LoadArchive(label string) (Archive, error)
}

type snapshot struct {
	baseline  string
	manifests string
	data      string
}

func (s *snapshot) CreateArchive(label string) (Archive, error) {

	// the baseline snapshot file
	bfile := ""
	if len(s.baseline) > 0 {
		bfile = filepath.Join(s.baseline, label) + ".csv"
	}

	// the snapshot file
	mfile := filepath.Join(s.manifests, label) + ".csv"

	// it's an error if the snapshot already exists
	_, err := os.Stat(mfile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		return nil, os.ErrExist
	}

	a := archive{
		baseline: bfile,
		manifest: mfile,
		data:     s.data,
	}
	return &a, nil
}

func (s *snapshot) LoadArchive(label string) (Archive, error) {

	// the snapshot file
	mfile := filepath.Join(s.manifests, label) + ".csv"

	// it's an error if the snapshot doesn't exist
	_, err := os.Stat(mfile)
	if err != nil {
		return nil, err
	}

	a := archive{
		baseline: "",
		manifest: mfile,
		data:     s.data,
	}
	return &a, nil
}
