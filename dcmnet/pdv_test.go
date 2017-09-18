package dcmnet

import (
	"testing"
)

func TestCommandAndLast(t *testing.T) {
	var flags PDVFlags
	setType := func(pdvType PDVType) func() {
		return func() {
			flags.SetType(pdvType)
		}
	}
	setLast := func(last bool) func() {
		return func() {
			flags.SetLast(last)
		}
	}

	// Unlike a normal table-driven test, this one keeps state and needs to
	// execute in order.  It makes all 4 possible transitions from all 4
	// possible states (the detail is justified by the relatively tricky
	// implementation).
	for i, test := range []struct {
		mod     func()
		expType PDVType
		expLast bool
	}{
		{func() {}, Data, false},
		{setType(Command), Command, false},
		{setLast(true), Command, true},
		{setLast(false), Command, false},
		{setType(Data), Data, false},
		{setLast(true), Data, true},
		{setType(Command), Command, true},
		{setType(Data), Data, true},
		{setLast(false), Data, false},
	} {
		test.mod()
		if flags.GetType() != test.expType || flags.IsLast() != test.expLast {
			t.Fatalf("%d: unexpected flags: %s", i, flags)
		}
		test.mod() // 2nd time should be a no-op
		if flags.GetType() != test.expType || flags.IsLast() != test.expLast {
			t.Fatalf("%d: unexpected flags: %s", i, flags)
		}
	}
}

func TestReadOnePDV(t *testing.T) {
	decoder := pdvDecoder(bufpdv(1, Command, true, "command"))
	expectNextPDV(t, decoder, 1, Command, true, "command")
	expectNoMorePDVs(t, decoder)
}

func TestReadMultiplePDVs(t *testing.T) {
	decoder := pdvDecoder(
		bufpdv(1, Command, true, "command"),
		bufpdv(2, Data, false, "data1"),
		bufpdv(3, Data, true, "data2"))

	expectNextPDV(t, decoder, 1, Command, true, "command")
	expectNextPDV(t, decoder, 2, Data, false, "data1")
	expectNextPDV(t, decoder, 3, Data, true, "data2")
	expectNoMorePDVs(t, decoder)
}

func TestNoPDVs(t *testing.T) {
	decoder := pdvDecoder()
	expectNoMorePDVs(t, decoder)
}

func TestDrainFirstPDVWhenAskedForSecond(t *testing.T) {
	decoder := pdvDecoder(
		bufpdv(1, Command, true, "command"),
		bufpdv(2, Data, false, "data"))
	// not reading value...
	decoder.NextPDV()
	expectNextPDV(t, decoder, 2, Data, false, "data")
	expectNoMorePDVs(t, decoder)
}

func expectNextPDV(
	t *testing.T,
	decoder PDVDecoder,
	context PCID,
	tipe PDVType,
	last bool,
	content string,
) {
	pdv, err := decoder.NextPDV()
	if err != nil {
		t.Fatal(err)
	}
	if pdv == nil {
		t.Fatal("didn't get expected pdv")
	}

	if pdv.Context != context || pdv.GetType() != tipe || pdv.IsLast() != last {
		t.Fatalf("got unexpected pdv: %s", pdv)
	}
	if got := toString(pdv.Data); got != content {
		t.Fatalf("got unexpected pdv content: %q", got)
	}
}

func expectNoMorePDVs(t *testing.T, decoder PDVDecoder) {
	pdv, err := decoder.NextPDV()
	if err != nil {
		t.Fatal(err)
	}
	if pdv != nil {
		t.Fatalf("unexpected pdv: %q", pdv)
	}
}

func pdvDecoder(pdvs ...interface{}) PDVDecoder {
	b := bufcat(pdvs...)
	return NewPDVDecoder(&b)
}
