package storage

type StorageLocation interface {
	BytesURL() string
	OutboardBytesURL() string
	String() string
	ProviderMessage() []byte
	Type() int
	Parts() []string
	BinaryParts() [][]byte
	Expiry() int64
	SetProviderMessage(msg []byte)
	SetType(t int)
	SetParts(p []string)
	SetBinaryParts(bp [][]byte)
	SetExpiry(e int64)
}

func (s *StorageLocationImpl) Type() int {
	return s.kind
}

func (s *StorageLocationImpl) Parts() []string {
	return s.parts
}

func (s *StorageLocationImpl) BinaryParts() [][]byte {
	return s.binaryParts
}

func (s *StorageLocationImpl) Expiry() int64 {
	return s.expiry
}

func (s *StorageLocationImpl) SetType(t int) {
	s.kind = t
}

func (s *StorageLocationImpl) SetParts(p []string) {
	s.parts = p
}

func (s *StorageLocationImpl) SetBinaryParts(bp [][]byte) {
	s.binaryParts = bp
}

func (s *StorageLocationImpl) SetExpiry(e int64) {
	s.expiry = e
}

func (s *StorageLocationImpl) SetProviderMessage(msg []byte) {
	s.providerMessage = msg
}

func (s *StorageLocationImpl) ProviderMessage() []byte {
	return s.providerMessage
}

func NewStorageLocation(Type int, Parts []string, Expiry int64) StorageLocation {
	return &StorageLocationImpl{
		kind:   Type,
		parts:  Parts,
		expiry: Expiry,
	}
}

type StorageLocationImpl struct {
	kind            int
	parts           []string
	binaryParts     [][]byte
	expiry          int64
	providerMessage []byte
}
