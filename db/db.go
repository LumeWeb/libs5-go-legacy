package db

type KVStore interface {
	Open() error
	Close() error
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
	Bucket(prefix string) (KVStore, error)
}

type Cache interface {
	Get(key []byte) ([]byte, bool)
	Put(key []byte, value []byte)
	Delete(key []byte)
}
