package kv

import (
	"context"
	"errors"
	"go.etcd.io/etcd/client/v3"
	"strings"
	"time"
)

var _ KVStore = (*EtcdKVStore)(nil)

const defaultTimeout = 10 * time.Second

// EtcdKVStore is an implementation of the KVStore interface using etcd.
type EtcdKVStore struct {
	client     *clientv3.Client
	prefix     string
	bucketName string
	timeout    time.Duration
	cache      Cache
}

// NewEtcdKVStore creates a new instance of EtcdKVStore.
func NewEtcdKVStore(client *clientv3.Client, prefix string, cache Cache, timeout time.Duration) *EtcdKVStore {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &EtcdKVStore{
		client:  client,
		prefix:  prefix,
		timeout: timeout,
		cache:   cache,
	}
}

// Open establishes a connection to the etcd cluster.
func (e *EtcdKVStore) Open() error {
	// The connection is already established in NewEtcdKVStore,
	// so this method can be left empty.
	return nil
}

// Close closes the connection to the etcd cluster.
func (e *EtcdKVStore) Close() error {
	return e.client.Close()
}

// Get retrieves the value associated with the given key.
func (e *EtcdKVStore) Get(key []byte) ([]byte, error) {
	// Check if the value exists in the cache
	if e.cache != nil {
		if val, ok := e.cache.Get(key); ok {
			return val, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	resp, err := e.client.Get(ctx, e.getKey(key))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	value := resp.Kvs[0].Value
	if e.cache != nil {
		e.cache.Put(key, value)
	}

	return value, nil
}

// Put stores the given key-value pair.
func (e *EtcdKVStore) Put(key []byte, value []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	_, err := e.client.Put(ctx, e.getKey(key), string(value))
	if err != nil {
		return err
	}

	// Update the cache with the new value
	if e.cache != nil {
		e.cache.Put(key, value)
	}

	return nil
}

// Delete removes the value associated with the given key.
func (e *EtcdKVStore) Delete(key []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	_, err := e.client.Delete(ctx, e.getKey(key))
	if err != nil {
		return err
	}

	// Remove the key from the cache
	if e.cache != nil {
		e.cache.Delete(key)
	}

	return nil
}

// Bucket creates a new bucket with the given prefix.
func (e *EtcdKVStore) Bucket(prefix string) (KVStore, error) {
	if strings.Contains(prefix, "/") {
		return nil, errors.New("bucket name cannot contain '/'")
	}
	return &EtcdKVStore{
		client:     e.client,
		prefix:     e.prefix + "/" + prefix,
		bucketName: prefix,
		timeout:    e.timeout,
		cache:      e.cache,
	}, nil
}

// getKey constructs the full key path for etcd.
func (e *EtcdKVStore) getKey(key []byte) string {
	return e.prefix + "/" + e.bucketName + "/" + string(key)
}
