package rootchain

import "github.com/0xPolygon/polygon-edge/types"

type PayloadType uint8

const (
	ValidatorSetPayload PayloadType = iota
)

type Payload interface {
	Get() (PayloadType, []byte)
}

type SAM struct {
	Hash      types.Hash // unique hash of the event
	Signature []byte     // validator signature

	Event
}

type VerifiedSAM []SAM

func (v VerifiedSAM) Signatures() (signatures [][]byte) {
	signatures = make([][]byte, len(v))

	for index, m := range v {
		signatures[index] = m.Signature
	}

	return
}
