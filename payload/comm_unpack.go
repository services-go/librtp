package payload

import (
	"errors"
	"github.com/services-go/librtp/rtp"
)

type RtpCommUnpack struct {
	RtpPayloadHelper
}

func (up *RtpCommUnpack) Init(h RtpPayload, cbparam interface{}) {
	up.handler = h
	up.cbparam = cbparam
	up.flags = -1
}

func (up *RtpCommUnpack) Input(packet []byte, bytes int) (int, error) {
	var pkt rtp.RtpPacket
	err := rtp.RtpPacketDeserialize(&pkt, packet, bytes)
	if err != nil {
		return -1, err
	}
	if pkt.PayloadLen < 1 {
		return -1, errors.New("rtp payload len < 1.")
	}
	up.handler.Handle(up.cbparam, pkt.Payload, pkt.PayloadLen, pkt.Header.Timestamp, 0)
	return 1, nil
}
