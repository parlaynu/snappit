package ops

import "os"

// ItemState represents the state of the file system object compared
// to the last time it was observed. This is determined by the OpCombineStreams
// operator by comparing the state of the live object to the state of the
// object recorded in a manifest or a reference file system.
type EntryStatus int

const (
	StatusOk EntryStatus = iota
	StatusError
	StatusNew
	StatusModified
	StatusNotInManifest
	StatusNotInFilesystem
	StatusRemoved
	StatusRestored
)

func EntryStatusName(status EntryStatus) string {
	statusToName := map[EntryStatus]string{
		StatusOk:              "Ok",
		StatusError:           "Error",
		StatusNew:             "New",
		StatusModified:        "Modified",
		StatusNotInManifest:   "NotInManifest",
		StatusNotInFilesystem: "NotInFilesystem",
		StatusRemoved:         "Removed",
		StatusRestored:        "Restored",
	}
	return statusToName[status]
}

// ItemInfo represents the meta-data associate with a file system object
// as well as the application required state and status flags. This
// structure is passed between operators to help them determine if their
// specific operation needs to be performed or can be skipped.
type EntryInfo struct {
	Status  EntryStatus
	Path    string
	Hash    string
	Size    int64
	ModTime int64
	Mode    os.FileMode
	Error   error
}
