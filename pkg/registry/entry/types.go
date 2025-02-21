package entry

type SignedRegistryEntry interface {
	PK() []byte
	Revision() uint64
	Data() []byte
	Signature() []byte
	SetPK(pk []byte)
	SetRevision(revision uint64)
	SetData(data []byte)
	SetSignature(signature []byte)
	Verify() bool
}
type RegistryEntry interface {
	Sign() SignedRegistryEntry
}
