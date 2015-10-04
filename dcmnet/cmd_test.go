package dcmnet

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestCommandFieldReqOrRsp(t *testing.T) {
	RegisterTestingT(t)

	Expect(CStoreReq.IsReq()).To(BeTrue())
	Expect(CStoreReq.IsRsp()).To(BeFalse())

	Expect(CStoreRsp.IsReq()).To(BeFalse())
	Expect(CStoreRsp.IsRsp()).To(BeTrue())

	Expect(NGetReq.IsReq()).To(BeTrue())
	Expect(NGetReq.IsRsp()).To(BeFalse())

	Expect(NGetRsp.IsReq()).To(BeFalse())
	Expect(NGetRsp.IsRsp()).To(BeTrue())
}

func TestCommandFieldToggleReqRsp(t *testing.T) {
	RegisterTestingT(t)

	Expect(CStoreReq.GetReq()).To(Equal(CStoreReq))
	Expect(CStoreReq.GetRsp()).To(Equal(CStoreRsp))

	Expect(CStoreRsp.GetReq()).To(Equal(CStoreReq))
	Expect(CStoreRsp.GetRsp()).To(Equal(CStoreRsp))
}

func TestCommandDataSetType(t *testing.T) {
	RegisterTestingT(t)

	Expect(CommandHasNoDataSet.HasDataset()).To(BeFalse())
	Expect(CommandHasDataSet.HasDataset()).To(BeTrue())
	Expect(CommandDataSetType(1).HasDataset()).To(BeTrue())
}
