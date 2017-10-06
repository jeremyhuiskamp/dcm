package dcmnet

import (
	"testing"
)

func TestCommandFieldReqOrRsp(t *testing.T) {
	assertReq := func(what CommandField, isReq bool) {
		if what.IsReq() != isReq {
			t.Errorf("expected %s to be request? %t", what, isReq)
		}
		if what.IsRsp() == isReq {
			t.Errorf("expected %s to be response? %t", what, !isReq)
		}
	}

	assertReq(CStoreReq, true)
	assertReq(CStoreRsp, false)
	assertReq(NGetReq, true)
	assertReq(NGetRsp, false)
}

func TestCommandFieldToggleReqRsp(t *testing.T) {
	assertPair := func(req, rsp CommandField) {
		if req.GetReq() != req {
			t.Errorf("%s should be it's own request", req)
		}
		if req.GetRsp() != rsp {
			t.Errorf("%s should be the response to %s", rsp, req)
		}
		if rsp.GetReq() != req {
			t.Errorf("%s should be the request for %s", req, rsp)
		}
		if rsp.GetRsp() != rsp {
			t.Errorf("%s should be it's own response", rsp)
		}
	}

	assertPair(CStoreReq, CStoreRsp)
	assertPair(NGetReq, NGetRsp)
}

func TestCommandDataSetType(t *testing.T) {
	assertDataSet := func(typ CommandDataSetType, exp bool) {
		if typ.HasDataset() != exp {
			t.Errorf("expected %X to have dataset? %t",
				typ, exp)
		}
	}

	assertDataSet(CommandHasNoDataSet, false)
	assertDataSet(CommandHasDataSet, true)
	assertDataSet(CommandDataSetType(1), true)
}
