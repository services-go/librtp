package payload

import (
	"github.com/pkg/errors"
	"github.com/services-go/librtp/rtp"
)

const (
	N_AU_HEADER = 4
)

type RtpPackMpeg4Generic struct {
	pkt     rtp.RtpPacket
	handler RtpPayload
	cbparam interface{}
	size    int
}

func (p *RtpPackMpeg4Generic) Init(size int, payload uint8, seq uint16, ssrc uint32, handler RtpPayload, cbparam interface{}) {
	p.handler = handler
	p.size = size
	p.cbparam = cbparam

	p.pkt.Header.Version = rtp.RtpVersion
	p.pkt.Header.PayloadType = payload
	p.pkt.Header.SequenceNumber = seq
	p.pkt.Header.SSRC = ssrc
}

// destroy RTP Packer
func (p *RtpPackMpeg4Generic) Destroy() {
}

func (p *RtpPackMpeg4Generic) GetInfo() (seq uint16, timestamp uint32) {
	return p.pkt.Header.SequenceNumber, p.pkt.Header.Timestamp
}

func (p *RtpPackMpeg4Generic) Input(data []byte, bytes int, timestamp uint32) error {
	p.pkt.Header.Timestamp = timestamp //(uint32_t)time * KHz; // ms -> 90KHZ
	ptr := data
	if ptr[0] == 0xFF && (ptr[1]&0xF0) == 0xF0 && bytes > 7 {
		// skip ADTS header
		if (int(ptr[3]&0x03)<<11)|(int(ptr[4])<<3|int((ptr[5]>>5)&0x07)) != bytes {
			return errors.New("error ADTS header.")
		}
		ptr = ptr[7:]
		bytes -= 7
	}

	var header [4]byte
	var n int
	for size := bytes; bytes > 0; p.pkt.Header.SequenceNumber++ {
		// 3.3.6. High Bit-rate AAC
		// SDP fmtp: sizeLength=13; indexLength=3; indexDeltaLength = 3;
		header[0] = 0
		header[1] = 16
		header[2] = byte(size >> 5)
		header[3] = byte(size&0x1f) << 3

		p.pkt.Payload = ptr
		p.pkt.PayloadLen = p.size - N_AU_HEADER - rtp.RtpFixedHeader
		if bytes+N_AU_HEADER+rtp.RtpFixedHeader <= p.size {
			p.pkt.PayloadLen = bytes
		}
		ptr = ptr[p.pkt.PayloadLen:]
		bytes -= p.pkt.PayloadLen

		n = rtp.RtpFixedHeader + N_AU_HEADER + p.pkt.PayloadLen
		rtpb := p.handler.Alloc(p.cbparam, n)
		if rtpb == nil {
			return errors.New("alloc rtp buffer failed.")
		}

		// Marker (M) bit: The M bit is set to 1 to indicate that the RTP packet
		// payload contains either the final fragment of a fragmented Access
		// Unit or one or more complete Access Units
		p.pkt.Header.Marker = 0
		if bytes == 0 {
			p.pkt.Header.Marker = 1
		}
		var err error
		n, err = rtp.RtpPacketSerializeHeader(&p.pkt, rtpb, n)
		if err != nil {
			return err
		}
		if n != rtp.RtpFixedHeader {
			return errors.New("rtp packet serialize header failed.")
		}

		copy(rtpb[n:], header[:N_AU_HEADER])
		copy(rtpb[n+N_AU_HEADER:], p.pkt.Payload[:p.pkt.PayloadLen])
		p.handler.Handle(p.cbparam, rtpb, n+N_AU_HEADER+p.pkt.PayloadLen, p.pkt.Header.Timestamp, 0)
		p.handler.Free(p.cbparam, rtpb)
	}
	return nil
}
