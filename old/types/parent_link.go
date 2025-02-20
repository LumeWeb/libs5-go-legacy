package types

type ParentLinkType int

const (
	ParentLinkTypeUserIdentity ParentLinkType = 0x01
	ParentLinkTypeBoard        ParentLinkType = 0x05
	ParentLinkTypeBridgeUser   ParentLinkType = 0x0A
)

var ParentLinkTypeMap = map[ParentLinkType]string{
	ParentLinkTypeUserIdentity: "UserIdentity",
	ParentLinkTypeBoard:        "Board",
	ParentLinkTypeBridgeUser:   "BridgeUser",
}
