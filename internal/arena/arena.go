package arena

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Arena interface {
	Path() string

	Exists() bool
	Create() error

	Snapshots() ([]string, error)

	CreateSnapshot(name, baseline string) (Snapshot, error)
	LoadSnapshot(name string) (Snapshot, error)
}

type arena struct {
	root      string
	snapshots string
	data      string
}

func New(root string) (Arena, error) {
	a := arena{
		root:      root,
		snapshots: filepath.Join(root, "snapshots"),
		data:      filepath.Join(root, "data"),
	}
	return &a, nil
}

func (a *arena) Path() string {
	return a.root
}

func (a *arena) Exists() bool {
	info, err := os.Stat(a.root)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (a *arena) Create() error {

	err := os.MkdirAll(a.snapshots, 0770)
	if err != nil {
		return err
	}
	err = os.MkdirAll(a.data, 0770)
	if err != nil {
		return err
	}

	return nil
}

func (a *arena) Snapshots() ([]string, error) {
	var snaps []string

	entries, err := os.ReadDir(a.snapshots)
	if err != nil {
		return nil, err
	}

	// loop over the source files
	for _, entry := range entries {
		if entry.IsDir() {
			snaps = append(snaps, entry.Name())
		}
	}

	slices.SortFunc(snaps, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	return snaps, nil
}

func (a *arena) CreateSnapshot(name, baseline string) (Snapshot, error) {
	// the baseline is the reference snapshot - make sure it exists
	if len(baseline) > 0 {
		baseline = filepath.Join(a.snapshots, baseline)
		_, err := os.Stat(baseline)
		if err != nil {
			return nil, err
		}
	}

	// this is where this snapshot will write snapshots
	snapdir := filepath.Join(a.snapshots, name)
	err := os.MkdirAll(snapdir, 0770)
	if err != nil {
		return nil, err
	}

	s := snapshot{
		baseline:  baseline,
		manifests: snapdir,
		data:      a.data,
	}
	return &s, nil
}

func (a *arena) LoadSnapshot(name string) (Snapshot, error) {

	// this is where this snapshot will write snapshots
	snapdir := filepath.Join(a.snapshots, name)
	_, err := os.Stat(snapdir)
	if err != nil {
		return nil, err
	}

	s := snapshot{
		baseline:  "",
		manifests: snapdir,
		data:      a.data,
	}
	return &s, nil
}
