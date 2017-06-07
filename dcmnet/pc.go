package dcmnet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/jeremyhuiskamp/dcm/dcm"
)

// PCID == presentation context id
type PCID uint8

// PCResult is a presentation context result.
type PCResult uint8

//go:generate stringer -type PCResult
const (
	PCAcceptance                                    PCResult = 0
	PCUserRejection                                 PCResult = 1
	PCProviderRejectionNoReason                     PCResult = 2
	PCProviderRejectionAbstractSyntaxNotSupported   PCResult = 3
	PCProviderRejectionTransferSyntaxesNotSupported PCResult = 4
)

func (pcr PCResult) IsAcceptance() bool {
	return pcr == PCAcceptance
}

type PresentationContext struct {
	ID               PCID
	Result           PCResult
	AbstractSyntax   string
	TransferSyntaxes []dcm.TransferSyntax
}

func (pc PresentationContext) Write(dst io.Writer) {
	// item type
	// TODO wrong! differentiate between request and response
	// TODO this should be a uint8 and a reserved 0x0
	binary.Write(dst, binary.LittleEndian, uint16(0x20))

	buf := new(bytes.Buffer)

	// TODO: write the result!
	buf.WriteByte(byte(pc.ID))
	buf.WriteByte(0)
	buf.WriteByte(0)
	buf.WriteByte(0)

	// TODO use constants here and below
	binary.Write(buf, binary.LittleEndian, uint16(0x30))
	binary.Write(buf, binary.BigEndian, uint16(len(pc.AbstractSyntax)))
	buf.WriteString(pc.AbstractSyntax)

	for _, ts := range pc.TransferSyntaxes {
		binary.Write(buf, binary.LittleEndian, uint16(0x40))
		binary.Write(buf, binary.BigEndian, uint16(len(ts.UID())))
		buf.WriteString(ts.UID())
	}

	binary.Write(dst, binary.BigEndian, uint16(buf.Len()))
	dst.Write(buf.Bytes())
}

func readString(src io.Reader) (string, error) {
	bytes, err := ioutil.ReadAll(src)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (pc *PresentationContext) Read(src io.Reader) error {
	var buf [4]byte
	_, err := src.Read(buf[:])
	if err != nil {
		return err
	}
	pc.ID = PCID(buf[0])

	pc.Result = PCResult(buf[2])

	return EachItem(src, func(item *Item) (err error) {
		switch item.Type {
		case AbstractSyntax:
			pc.AbstractSyntax, err = readString(item.Data)
			if err != nil {
				return err
			}
		case TransferSyntax:
			ts, err := readString(item.Data)
			if err != nil {
				return err
			}
			pc.TransferSyntaxes = append(pc.TransferSyntaxes,
				dcm.GetTransferSyntax(ts))
		}

		return nil
	})
}

// Role selects whether the association requestor is an SCU, SCP or both for any
// given abstract syntax.
// The default value is SCU only (see DefaultRole below).
// See PS 3.7, D.3.3.4
type Role [2]uint8

func NewRole(scu, scp bool) (r Role) {
	r.SetSCU(scu)
	r.SetSCP(scp)
	return
}

func (r Role) IsSCU() bool {
	return r[0] != 0
}

func (r *Role) SetSCU(scu bool) {
	if scu {
		r[0] = 1
	} else {
		r[0] = 0
	}
}

func (r Role) IsSCP() bool {
	return r[1] != 0
}

func (r *Role) SetSCP(scp bool) {
	if scp {
		r[1] = 1
	} else {
		r[1] = 0
	}
}

func (r Role) String() string {
	return fmt.Sprintf("Role: scu=%t, scp=%t", r.IsSCU(), r.IsSCP())
}

// DefaultRole is the value that should be assumed if there is no role selection
// in the association request.  It is SCU but not SCP.
// Please do not change the values!
var DefaultRole = NewRole(true, false)

// TransferCapability is the ability to send a sop with a particular abstract
// syntax and transfer syntax.
type TransferCapability struct {
	AbstractSyntax string
	Role           Role
	// Not sure if this should be a slice or a single value.  When requesting
	// an association, multiple values may be used, but these could be described
	// by multiple transfer capabilities (the downside there being that there
	// can only be one role per abstract syntax so this would allow for
	// inconsistencies).
	// When accepting and using a transfer capability, there can only be one
	// syntax, so having multiples here is annoying.
	TransferSyntaxes []dcm.TransferSyntax
}

func NewTransferCapability(abstractSyntax string, transferSyntaxes ...dcm.TransferSyntax) TransferCapability {
	return TransferCapability{
		AbstractSyntax:   abstractSyntax,
		TransferSyntaxes: transferSyntaxes,
	}
}

// PresentationContexts combines requested and accepted presentation contexts
// to map between presentation context ids and transfer capabilities.
type PresentationContexts struct {
	Requested []PresentationContext
	Accepted  []PresentationContext
}

// TODO: these methods need to be retrofitted to support roles

func (pc *PresentationContexts) findAcceptedPC(f func(rq, ac PresentationContext) bool,
) (rq, ac *PresentationContext) {
	// could save a lot of work here by pre-matching the requests and accepts
	// by id
	for _, rqpc := range pc.Requested {
		for _, acpc := range pc.Accepted {
			if rqpc.ID == acpc.ID && acpc.Result.IsAcceptance() && f(rqpc, acpc) {
				return &rqpc, &acpc
			}
		}
	}

	return nil, nil
}

func (pc *PresentationContexts) FindAcceptedTCap(pcid PCID) *TransferCapability {
	rqpc, acpc := pc.findAcceptedPC(func(rq, ac PresentationContext) bool {
		return rq.ID == pcid
	})

	if rqpc == nil {
		return nil
	}

	return &TransferCapability{
		AbstractSyntax: rqpc.AbstractSyntax,
		// should only have one value:
		TransferSyntaxes: acpc.TransferSyntaxes,
	}
}

func overlap(one, two []dcm.TransferSyntax) bool {
	for _, tsone := range one {
		for _, tstwo := range two {
			if tsone.UID() == tstwo.UID() {
				return true
			}
		}
	}

	return false
}

func (pc *PresentationContexts) FindAcceptedPCID(tcap TransferCapability) *PCID {
	rqpc, _ := pc.findAcceptedPC(func(rq, ac PresentationContext) bool {
		return rq.AbstractSyntax == tcap.AbstractSyntax &&
			overlap(ac.TransferSyntaxes, tcap.TransferSyntaxes)
	})

	if rqpc == nil {
		return nil
	}

	return &rqpc.ID
}
