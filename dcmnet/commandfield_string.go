// generated by stringer -type CommandField; DO NOT EDIT

package dcmnet

import "fmt"

const _CommandField_name = "CStoreReqCGetReqCFindReqCMoveReqCEchoReqNEventReportReqNGetReqNSetReqNActionReqNCreateReqNDeleteReqCCancelReqRspMaskCStoreRspCGetRspCFindRspCMoveRspCEchoRspNEventReportRspNGetRspNSetRspNActionRspNCreateRspNDeleteRsp"

var _CommandField_map = map[CommandField]string{
	1:     _CommandField_name[0:9],
	16:    _CommandField_name[9:16],
	32:    _CommandField_name[16:24],
	33:    _CommandField_name[24:32],
	48:    _CommandField_name[32:40],
	256:   _CommandField_name[40:55],
	272:   _CommandField_name[55:62],
	288:   _CommandField_name[62:69],
	304:   _CommandField_name[69:79],
	320:   _CommandField_name[79:89],
	336:   _CommandField_name[89:99],
	4095:  _CommandField_name[99:106],
	32768: _CommandField_name[106:116],
	32769: _CommandField_name[116:125],
	32784: _CommandField_name[125:132],
	32800: _CommandField_name[132:140],
	32801: _CommandField_name[140:148],
	32816: _CommandField_name[148:156],
	33024: _CommandField_name[156:171],
	33040: _CommandField_name[171:178],
	33056: _CommandField_name[178:185],
	33072: _CommandField_name[185:195],
	33088: _CommandField_name[195:205],
	33104: _CommandField_name[205:215],
}

func (i CommandField) String() string {
	if str, ok := _CommandField_map[i]; ok {
		return str
	}
	return fmt.Sprintf("CommandField(%d)", i)
}
