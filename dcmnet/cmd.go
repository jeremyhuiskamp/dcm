package dcmnet

// CommandField is a value that can appear in (0000,0100)
type CommandField uint16

//go:generate stringer -type CommandField
const (
	CStoreReq CommandField = 0x0001
	CStoreRsp CommandField = 0x8001
	CGetReq   CommandField = 0x0010
	CGetRsp   CommandField = 0x8010
	CFindReq  CommandField = 0x0020
	CFindRsp  CommandField = 0x8020
	CMoveReq  CommandField = 0x0021
	CMoveRsp  CommandField = 0x8021
	CEchoReq  CommandField = 0x0030
	CEchoRsp  CommandField = 0x8030

	NEventReportReq CommandField = 0x0100
	NEventReportRsp CommandField = 0x8100
	NGetReq         CommandField = 0x0110
	NGetRsp         CommandField = 0x8110
	NSetReq         CommandField = 0x0120
	NSetRsp         CommandField = 0x8120
	NActionReq      CommandField = 0x0130
	NActionRsp      CommandField = 0x8130
	NCreateReq      CommandField = 0x0140
	NCreateRsp      CommandField = 0x8140
	NDeleteReq      CommandField = 0x0150
	NDeleteRsp      CommandField = 0x8150

	CCancel CommandField = 0x0FFF

	ReqRspMask CommandField = 0x8000
)

func (cf CommandField) IsReq() bool {
	return cf&ReqRspMask == 0
}

func (cf CommandField) IsRsp() bool {
	return cf&ReqRspMask != 0
}

func (cf CommandField) GetReq() CommandField {
	return cf & ^ReqRspMask
}

func (cf CommandField) GetRsp() CommandField {
	return cf | ReqRspMask
}

// CommandDataSetType is a value that can appear in (0000,0800).
type CommandDataSetType uint16

const (
	// CommandHasNoDataSet is the only value that indicates that the command
	// has no data set
	CommandHasNoDataSet CommandDataSetType = 0x0101
	// CommandHasData is the historical value that indicates that the command
	// has a data set.  However, any value that is not CommandHasNoDataSet
	// (0x0101) means the same thing.
	CommandHasDataSet CommandDataSetType = 0x0102
)

func (cdst CommandDataSetType) HasDataset() bool {
	return cdst != CommandHasNoDataSet
}
