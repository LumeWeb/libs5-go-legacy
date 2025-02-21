package provider

import (
	"context"
	"go.lumeweb.com/libs5-go/pkg/crypto"
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"go.lumeweb.com/libs5-go/pkg/fs"
)

type S5APIProvider interface {
	// EnsureInitialized blocks until the S5 API is initialized and ready to be used
	EnsureInitialized(ctx context.Context) error

	// UploadBlob uploads a small blob of bytes
	// Returns the Raw CID of the uploaded raw file blob
	// Max size is 10 MiB, use UploadRawFile for larger files
	UploadBlob(ctx context.Context, data []byte) (encoding.BlobCID, error)

	// UploadBlobWithStream uploads a raw file
	// Returns the Raw CID of the uploaded raw file blob
	// Does not have a file size limit and can handle large files efficiently
	UploadBlobWithStream(ctx context.Context, hash encoding.Blob, size int64, openRead fs.OpenReadFunction) (encoding.BlobCID, error)

	// DownloadBlob downloads a full file blob to memory
	// Only use this if blobs are smaller than 1 MB
	DownloadBlob(ctx context.Context, hash encoding.Blob, route *encoding.Route) ([]byte, error)

	// DownloadBlobSlice downloads a slice of a blob to memory
	// From start (inclusive) to end (exclusive)
	DownloadBlobSlice(ctx context.Context, hash encoding.Blob, start, end int64, route *encoding.Route) ([]byte, error)

	// PinHash pins a hash to ensure it remains available
	PinHash(ctx context.Context, hash encoding.Blob) error

	// UnpinHash unpins a previously pinned hash
	UnpinHash(ctx context.Context, hash encoding.Blob) error

	// RegistryGet retrieves a registry entry
	RegistryGet(ctx context.Context, pk []byte, route *encoding.Route) (*encoding.SignedRegistryEntry, error)

	// RegistryListen listens for registry changes
	RegistryListen(ctx context.Context, pk []byte, route *encoding.Route) (<-chan encoding.SignedRegistryEntry, error)

	// RegistrySet sets a registry entry
	RegistrySet(ctx context.Context, sre encoding.SignedRegistryEntry, route *encoding.Route) error

	// StreamSubscribe subscribes to a stream
	StreamSubscribe(ctx context.Context, pk []byte, afterTimestamp, beforeTimestamp *int64, route *encoding.Route) (<-chan encoding.SignedStreamMessage, error)

	// StreamPublish publishes a message to a stream
	StreamPublish(ctx context.Context, msg encoding.SignedStreamMessage, route *encoding.Route) error

	// Crypto returns the crypto implementation
	Crypto() crypto.CryptoImplementation
}
