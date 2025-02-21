package libs5_go

import (
	"context"
	"io"
)

// OpenReadFunction is a function that returns a reader for a file or blob

// S5APIProvider defines the interface for interacting with the S5 API
type S5APIProvider interface {
	// EnsureInitialized blocks until the S5 API is initialized and ready to be used
	EnsureInitialized(ctx context.Context) error

	// UploadBlob uploads a small blob of bytes
	// Returns the Raw CID of the uploaded raw file blob
	// Max size is 10 MiB, use UploadRawFile for larger files
	UploadBlob(ctx context.Context, data []byte) (BlobCID, error)

	// UploadBlobWithStream uploads a raw file
	// Returns the Raw CID of the uploaded raw file blob
	// Does not have a file size limit and can handle large files efficiently
	UploadBlobWithStream(ctx context.Context, hash Multihash, size int64, openRead OpenReadFunction) (BlobCID, error)

	// DownloadBlob downloads a full file blob to memory
	// Only use this if blobs are smaller than 1 MB
	DownloadBlob(ctx context.Context, hash Multihash, route *Route) ([]byte, error)

	// DownloadBlobSlice downloads a slice of a blob to memory
	// From start (inclusive) to end (exclusive)
	DownloadBlobSlice(ctx context.Context, hash Multihash, start, end int64, route *Route) ([]byte, error)

	// PinHash pins a hash to ensure it remains available
	PinHash(ctx context.Context, hash Multihash) error

	// UnpinHash unpins a previously pinned hash
	UnpinHash(ctx context.Context, hash Multihash) error

	// RegistryGet retrieves a registry entry
	RegistryGet(ctx context.Context, pk []byte, route *Route) (*SignedRegistryEntry, error)

	// RegistryListen listens for registry changes
	RegistryListen(ctx context.Context, pk []byte, route *Route) (<-chan SignedRegistryEntry, error)

	// RegistrySet sets a registry entry
	RegistrySet(ctx context.Context, sre SignedRegistryEntry, route *Route) error

	// StreamSubscribe subscribes to a stream
	StreamSubscribe(ctx context.Context, pk []byte, afterTimestamp, beforeTimestamp *int64, route *Route) (<-chan SignedStreamMessage, error)

	// StreamPublish publishes a message to a stream
	StreamPublish(ctx context.Context, msg SignedStreamMessage, route *Route) error

	// Crypto returns the crypto implementation
	Crypto() CryptoImplementation
}

// These are placeholder type definitions that would need to be implemented
type (
	BlobCID             []byte
	Multihash           []byte
	Route               struct{}
	SignedRegistryEntry struct{}
	SignedStreamMessage struct{}
	KeyPairEd25519      struct {
		PublicKey  []byte
		PrivateKey []byte
	}
)
