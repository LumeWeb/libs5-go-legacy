package types

type StorageLocationType int

const (
	StorageLocationTypeArchive StorageLocationType = 0
	StorageLocationTypeFile    StorageLocationType = 3
	StorageLocationTypeFull    StorageLocationType = 5
	StorageLocationTypeBridge  StorageLocationType = 7
)

var StorageLocationTypeMap = map[StorageLocationType]string{
	StorageLocationTypeArchive: "Archive",
	StorageLocationTypeFile:    "File",
	StorageLocationTypeFull:    "Full",
	StorageLocationTypeBridge:  "Bridge",
}
