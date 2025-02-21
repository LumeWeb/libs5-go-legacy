package kv

import (
	"errors"

	"go.etcd.io/bbolt"
)

var _ KVStore = (*BboltDBKVStore)(nil)

var (
	ErrGetRoot    = errors.New("cannot get from root")
	ErrDeleteRoot = errors.New("cannot delete from root")
	ErrorPutRoot  = errors.New("cannot put from root")
)

type BboltDBKVStore struct {
	db         *bbolt.DB
	bucket     *bbolt.Bucket
	bucketName string
	root       bool
	dbPath     string
	cache      Cache
}

func (b *BboltDBKVStore) Open() error {
	if b.root && b.db == nil {
		db, err := bbolt.Open(b.dbPath, 0666, nil)
		if err != nil {
			return err
		}
		b.db = db
	}

	if len(b.bucketName) > 0 {
		err := b.db.Update(func(txn *bbolt.Tx) error {
			var bucket *bbolt.Bucket
			var err error

			if b.bucket == nil {
				bucket, err = txn.CreateBucketIfNotExists([]byte(b.bucketName))
				if err != nil {
					return err
				}
			} else {
				bucket, err = b.bucket.CreateBucketIfNotExists([]byte(b.bucketName))
				if err != nil {
					return err
				}
			}

			b.bucket = bucket
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BboltDBKVStore) Close() error {
	if b.root && b.db != nil {
		err := b.db.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BboltDBKVStore) Get(key []byte) ([]byte, error) {
	if b.root {
		return nil, ErrGetRoot
	}

	// Check if the value exists in the cache
	if b.cache != nil {
		if val, ok := b.cache.Get(key); ok {
			return val, nil
		}
	}

	var val []byte
	err := b.db.View(func(txn *bbolt.Tx) error {
		bucket := txn.Bucket([]byte(b.bucketName))
		val = bucket.Get(key)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if val != nil && b.cache != nil {
		// Add the value to the cache
		b.cache.Put(key, val)
	}

	return val, nil
}

func (b *BboltDBKVStore) Put(key []byte, value []byte) error {
	if b.root {
		return ErrorPutRoot
	}

	err := b.db.Update(func(txn *bbolt.Tx) error {
		bucket := txn.Bucket([]byte(b.bucketName))
		err := bucket.Put(key, value)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Update the cache with the new value
	if b.cache != nil {
		b.cache.Put(key, value)
	}

	return nil
}

func (b *BboltDBKVStore) Delete(key []byte) error {
	if b.root {
		return ErrDeleteRoot
	}

	err := b.db.Update(func(txn *bbolt.Tx) error {
		bucket := txn.Bucket([]byte(b.bucketName))
		err := bucket.Delete(key)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Remove the key from the cache
	if b.cache != nil {
		b.cache.Delete(key)
	}

	return nil
}

func (b *BboltDBKVStore) Bucket(prefix string) (KVStore, error) {
	return &BboltDBKVStore{
		db:         b.db,
		bucket:     b.bucket,
		bucketName: prefix,
		root:       false,
		cache:      b.cache,
	}, nil
}

func NewBboltDBKVStore(dbPath string, cache Cache) *BboltDBKVStore {
	return &BboltDBKVStore{
		dbPath: dbPath,
		root:   true,
		cache:  cache,
	}
}
