package payload

import (
	"github.com/pkg/errors"
	"github.com/services-go/librtp/rtp"
)

type CommPack struct {
	handler RtpPayload
	cbparam interface{}
	pkt     rtp.RtpPacket
	size    int
}

func (p CommPack) Create(size int, payload uint8, seq uint16, ssrc uint32, handler RtpPayload, cbparam interface{}) RtpPayloadPacker {
	pack := &CommPack{}
	pack.cbparam = cbparam
	pack.size = size

	pack.pkt.Header.Version = rtp.RtpVersion
	pack.pkt.Header.PayloadType = payload
	pack.pkt.Header.SequenceNumber = seq
	pack.pkt.Header.SSRC = ssrc
	return pack
}

// destroy RTP Packer
func (p *CommPack) Destroy() {
}

func (p *CommPack) GetInfo() (seq uint16, timestamp uint32) {
	return p.pkt.Header.SequenceNumber, p.pkt.Header.Timestamp
}

// PS/H.264 Elementary Stream to RTP Packet
// @param[in] packer
// @param[in] data stream data
// @param[in] bytes stream length in bytes
// @param[in] time stream UTC time
// @return 0-ok, ENOMEM-alloc failed, <0-failed
func (p *CommPack) Input(data []byte, bytes int, timestamp uint32) error {
	if p.pkt.Payload != nil {
		return errors.New("not first packet.")
	}

	if p.pkt.Header.Timestamp == timestamp {
		return errors.New("error timestamp.")
	}

	p.pkt.Header.Timestamp = timestamp // (uint32_t)time * packer->frequency / 1000; // ms -> 8KHZ
	p.pkt.Header.Marker = 0            // marker bit alway 0

	var n int
	var r interface{}
	for ptr := data; bytes > 0; p.pkt.Header.SequenceNumber++ {
		p.pkt.Payload = ptr[:]
		p.pkt.PayloadLen = p.size - rtp.RtpFixedHeader
		if (bytes + rtp.RtpFixedHeader) <= p.size {
			p.pkt.PayloadLen = bytes
		}

		ptr = ptr[p.pkt.PayloadLen:]
		bytes -= p.pkt.PayloadLen
		n = rtp.RtpFixedHeader + p.pkt.PayloadLen
		r = p.handler.Alloc(p.cbparam, n)
		if r == nil {
			return errors.New("rtp_pack alloc failed.")
		}

	}
}
