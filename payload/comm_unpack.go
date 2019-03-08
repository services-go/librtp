package payload

import (
	"errors"
	"github.com/services-go/librtp/rtp"
)

type RtpUnpack struct {
	RtpPayloadHelper
}

func (up RtpUnpack) Create(h RtpPayload, cbparam interface{}) RtpPayloadUnpacker {
	unpack := &RtpUnpack{
		RtpPayloadHelper{
			handler: h,
			cbparam: cbparam,
			flags:   -1,
		},
	}
	return unpack
}

func (up *RtpUnpack) Input(packet []byte, bytes int) error {
	var pkt rtp.RtpPacket
	err := rtp.RtpPacketDeserialize(&pkt, packet, bytes)
	if err != nil {
		return err
	}
	if pkt.PayloadLen < 1 {
		return errors.New("rtp payload len < 1.")
	}
	up.handler.Packet(up.cbparam, pkt.Payload, pkt.PayloadLen, pkt.Header.Timestamp, 0)
	return nil
}
