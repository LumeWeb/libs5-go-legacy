package storage

import "go.lumeweb.com/libs5-go/pkg/encoding"

type SignedStorageLocation interface {
	String() string
	NodeId() *encoding.NodeId
	Location() StorageLocation
}
