package metadata

import "github.com/vmihailenco/msgpack/v5"

type Metadata interface {
}
type SerializableMetadata interface {
	msgpack.CustomEncoder
	msgpack.CustomDecoder
}

type BaseMetadata struct {
	Type string `json:"type"`
}
