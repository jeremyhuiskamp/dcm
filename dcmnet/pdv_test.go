package dcmnet

import (
	. "gopkg.in/check.v1"
	"testing"
)

func TestPDV(t *testing.T) {
	Suite(&PDVTest{})
	TestingT(t)
}

type PDVTest struct{}

func (t *PDVTest) TestCommandVsData(c *C) {
	pdv := PDV{}

	c.Assert(pdv.GetType(), Equals, Data) // Data is default

	pdv.SetType(Command)
	c.Assert(pdv.GetType(), Equals, Command)

	pdv.SetType(Data)
	c.Assert(pdv.GetType(), Equals, Data)
}

func (t *PDVTest) TestLast(c *C) {
	pdv := PDV{}

	c.Assert(pdv.IsLast(), Equals, false) // Not Last is default

	pdv.SetLast(true)
	c.Assert(pdv.IsLast(), Equals, true)

	pdv.SetLast(false)
	c.Assert(pdv.IsLast(), Equals, false)
}

func (t *PDVTest) TestLastDoesntAffectCommand(c *C) {
	pdv := PDV{}
	
	c.Assert(pdv.GetType(), Equals, Data)
	pdv.SetLast(!pdv.IsLast())
	c.Assert(pdv.GetType(), Equals, Data)
	pdv.SetLast(!pdv.IsLast())
	c.Assert(pdv.GetType(), Equals, Data)
	
	pdv.SetType(Command)
	pdv.SetLast(!pdv.IsLast())
	c.Assert(pdv.GetType(), Equals, Command)
	pdv.SetLast(!pdv.IsLast())
	c.Assert(pdv.GetType(), Equals, Command)
	
	pdv.SetType(Data)
	pdv.SetLast(!pdv.IsLast())
	c.Assert(pdv.GetType(), Equals, Data)
	pdv.SetLast(!pdv.IsLast())
	c.Assert(pdv.GetType(), Equals, Data)
}

func (t *PDVTest) TestCommandDoesntAffectLast(c *C) {
	pdv := PDV{}
	
	c.Assert(pdv.IsLast(), Equals, false)
	pdv.SetType(Command)
	c.Assert(pdv.IsLast(), Equals, false)
	pdv.SetType(Data)
	c.Assert(pdv.IsLast(), Equals, false)
	
	pdv.SetLast(true)
	pdv.SetType(Command)
	c.Assert(pdv.IsLast(), Equals, true)
	pdv.SetType(Data)
	c.Assert(pdv.IsLast(), Equals, true)
	
	pdv.SetLast(false)
	pdv.SetType(Command)
	c.Assert(pdv.IsLast(), Equals, false)
	pdv.SetType(Data)
	c.Assert(pdv.IsLast(), Equals, false)
}
