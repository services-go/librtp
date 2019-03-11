package payload

import (
	"errors"
	"github.com/services-go/librtp/rtp"
)

type RtpCommPack struct {
	handler RtpPayload
	cbparam interface{}
	pkt     rtp.RtpPacket
	size    int
}

func (p *RtpCommPack) Init(size int, payload uint8, seq uint16, ssrc uint32, handler RtpPayload, cbparam interface{}) {
	p.cbparam = cbparam
	p.size = size

	p.pkt.Header.Version = rtp.RtpVersion
	p.pkt.Header.PayloadType = payload
	p.pkt.Header.SequenceNumber = seq
	p.pkt.Header.SSRC = ssrc
}

// destroy RTP Packer
func (p *RtpCommPack) Destroy() {
}

func (p *RtpCommPack) GetInfo() (seq uint16, timestamp uint32) {
	return p.pkt.Header.SequenceNumber, p.pkt.Header.Timestamp
}

// PS/H.264 Elementary Stream to RTP Packet
// @param[in] packer
// @param[in] data stream data
// @param[in] bytes stream length in bytes
// @param[in] time stream UTC time
// @return 0-ok, ENOMEM-alloc failed, <0-failed
func (p *RtpCommPack) Input(data []byte, bytes int, timestamp uint32) error {
	if p.pkt.Payload != nil {
		return errors.New("not first packet.")
	}

	if p.pkt.Header.Timestamp == timestamp {
		return errors.New("error timestamp.")
	}

	p.pkt.Header.Timestamp = timestamp // (uint32_t)time * packer->frequency / 1000; // ms -> 8KHZ
	p.pkt.Header.Marker = 0            // marker bit alway 0

	var n int
	var err error
	var rtpb []byte
	for ptr := data; bytes > 0; p.pkt.Header.SequenceNumber++ {
		p.pkt.Payload = ptr[:]
		p.pkt.PayloadLen = p.size - rtp.RtpFixedHeader
		if (bytes + rtp.RtpFixedHeader) <= p.size {
			p.pkt.PayloadLen = bytes
		}

		ptr = ptr[p.pkt.PayloadLen:]
		bytes -= p.pkt.PayloadLen
		n = rtp.RtpFixedHeader + p.pkt.PayloadLen
		rtpb = p.handler.Alloc(p.cbparam, n)
		if rtpb == nil {
			return errors.New("rtp_pack alloc failed.")
		}

		n, err = rtp.RtpPacketSerialize(&p.pkt, rtpb, n)
		if err != nil {
			return err
		}
		if n != rtp.RtpFixedHeader+p.pkt.PayloadLen {
			return errors.New("rtp packet serialize failed.")
		}

		p.handler.Handle(p.cbparam, rtpb, n, p.pkt.Header.Timestamp, 0)
		p.handler.Free(p.cbparam, rtpb)
	}

	return nil
}
