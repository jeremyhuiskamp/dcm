package dcmnet

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type AssociateRQAC struct {
	ProtocolVersion        int16
	CalledAE, CallingAE    string
	ApplicationContext     string
	PresentationContexts   []PresentationContext
	MaxPDULength           uint32
	ImplementationClassUID string
	ImplementationVersion  string
	MaxOperationsInvoked   uint16
	MaxOperationsPerformed uint16
	// TODO:
	// Roles
	// Extended Negotiation
	// Common Extended Negotiation
}

type AssociateRQ struct {
	AssociateRQAC

	// TODO:
	// User identity
}

type AssociateAC struct {
	AssociateRQAC

	// TODO:
	// User identity
}

// TODO: major amounts of error handling

func (rq AssociateRQAC) Write(dst io.Writer) {
	binary.Write(dst, binary.BigEndian, rq.ProtocolVersion)
	// padding:
	binary.Write(dst, binary.BigEndian, uint16(0))

	fmt.Fprintf(dst, "%-16s", rq.CalledAE)
	fmt.Fprintf(dst, "%-16s", rq.CallingAE)

	// reserved bytes:
	var reserved [32]byte
	dst.Write(reserved[:])

	for _, presentationContext := range rq.PresentationContexts {
		presentationContext.Write(dst)
	}
}

func readAE(src io.Reader) string {
	var bytes [16]byte
	io.ReadFull(src, bytes[:])
	return strings.TrimSpace(string(bytes[:]))
}

func (rq *AssociateRQAC) Read(src io.Reader) error {
	binary.Read(src, binary.BigEndian, &rq.ProtocolVersion)

	var padding uint16
	binary.Read(src, binary.BigEndian, &padding)

	rq.CalledAE = readAE(src)

	rq.CallingAE = readAE(src)

	var reserved [32]byte
	io.ReadFull(src, reserved[:])

	return EachItem(src, func(item *Item) error {
		switch item.Type {
		case ApplicationContext:
			s, err := readString(item.Data)
			if err != nil {
				return err
			}
			// hmm, weird, why can't I assign this directly from readString?
			rq.ApplicationContext = s

		case RequestPresentationContext, AcceptPresentationContext:
			pc := PresentationContext{}
			pc.Read(item.Data)
			rq.PresentationContexts = append(rq.PresentationContexts, pc)

		case UserInfo: // user info
			return readUserInfo(item.Data, rq)
		}

		// TODO: actual error handling above
		return nil
	})
}

func readUserInfo(src io.Reader, rqac *AssociateRQAC) error {
	return EachItem(src, func(item *Item) (err error) {
		switch item.Type {
		case MaxPDULength:
			return binary.Read(item.Data, binary.BigEndian, &rqac.MaxPDULength)

		case ImplementationClassUID:
			rqac.ImplementationClassUID, err = readString(item.Data)
			if err != nil {
				return err
			}

		case ImplementationVersion:
			rqac.ImplementationVersion, err = readString(item.Data)
			if err != nil {
				return err
			}

		case AsyncOperations:
			var values [4]byte
			_, err = io.ReadFull(item.Data, values[:])
			if err != nil {
				return err
			}

			rqac.MaxOperationsInvoked = binary.BigEndian.Uint16(values[:2])
			rqac.MaxOperationsPerformed = binary.BigEndian.Uint16(values[2:])
		}

		return nil
	})
}
