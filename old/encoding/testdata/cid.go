package testdata

import "encoding/hex"

var (
	RawCIDBytes         = []byte{0x26, 0x1f, 0x11, 0xaf, 0x66, 0xd2, 0x27, 0xde, 0x50, 0x1a, 0x40, 0x1b, 0xb0, 0x76, 0xd8, 0x0c, 0x5b, 0x50, 0x79, 0x51, 0xfa, 0xb2, 0xa7, 0xd7, 0xd6, 0x09, 0x6a, 0x64, 0xc8, 0x3a, 0x01, 0x1f, 0x6f, 0x60, 0x5e, 0x9f, 0x09}
	RawBase16CID        = "f261f11af66d227de501a401bb076d80c5b507951fab2a7d7d6096a64c83a011f6f605e9f09"
	RawBase32CID        = "beyprdl3g2it54ua2ian3a5wybrnva6kr7kzkpv6wbfvgjsb2aepw6yc6t4eq"
	RawBase58CID        = "z2H6yKf4s6awVkoiVJ4ARCZWLzX6eBhSaCkkqcjUCtmvqKcM4c5W"
	RawBase64CID        = "uJh8Rr2bSJ95QGkAbsHbYDFtQeVH6sqfX1glqZMg6AR9vYF6fCQ"
	RawCIDSize   uint32 = 630622

	MediaCIDBytes         = []byte{0xc5, 0x1f, 0x5f, 0x23, 0xf9, 0xe9, 0xbd, 0x46, 0xb2, 0xfa, 0x0c, 0xc5, 0xa4, 0x1e, 0x51, 0x8a, 0x2a, 0xd7, 0x5f, 0xc6, 0x83, 0x6c, 0x53, 0x22, 0xca, 0x7d, 0x2d, 0xbf, 0x0f, 0xd0, 0xe0, 0xd7, 0xbe, 0x9d}
	MediaBase16CID        = "fc51f5f23f9e9bd46b2fa0cc5a41e518a2ad75fc6836c5322ca7d2dbf0fd0e0d7be9d"
	MediaBase32CID        = "byupv6i7z5g6unmx2btc2ihsrrivnox6gqnwfgiwkpuw36d6q4dl35hi"
	MediaBase58CID        = "z5TTkenVbffNSgTcU4pkBcN2H1ZYctwLyQeLNEdr48tEpZHv"
	MediaBase64CID        = "uxR9fI_npvUay-gzFpB5RiirXX8aDbFMiyn0tvw_Q4Ne-nQ"
	MediaCIDSize   uint32 = 0

	ResolverCIDBytes            = []byte{0x25, 0xed, 0x2f, 0x66, 0xbf, 0xfa, 0xd8, 0x19, 0xa6, 0xbf, 0x22, 0x1d, 0x26, 0xee, 0x0f, 0xfe, 0x75, 0xe4, 0x8d, 0x15, 0x4f, 0x13, 0x76, 0x1e, 0xaa, 0xe5, 0x75, 0x89, 0x6f, 0x17, 0xdb, 0xda, 0x5f, 0xd3}
	ResolverBase16CID           = "f25ed2f66bffad819a6bf221d26ee0ffe75e48d154f13761eaae575896f17dbda5fd3"
	ResolverBase32CID           = "bexws6zv77lmbtjv7eiosn3qp7z26jdivj4jxmhvk4v2ys3yx3pnf7uy"
	ResolverBase58CID           = "zrjENDT9Doeok7pHUaojsYh5j3U1zKMudwTqZYNUftd8WCA"
	ResolverBase64CID           = "uJe0vZr_62BmmvyIdJu4P_nXkjRVPE3YequV1iW8X29pf0w"
	ResolverCIDSize      uint32 = 0
	ResolverData                = "5a591f6f05b96e1684cab18b410d7510a08392bf4b529a954e94636d6c9a5fc639cacd"
	ResolverDataBytes, _        = hex.DecodeString(ResolverData)
	ResolverDataSize     uint32 = 0
)
