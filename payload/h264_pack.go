// RFC6184 RTP Payload Format for H.264 Video
//
// 6.2. Single NAL Unit Mode (All receivers MUST support this mode)
//   packetization-mode media type parameter is equal to 0 or the packetization - mode is not present.
//   Only single NAL unit packets MAY be used in this mode.
//   STAPs, MTAPs, and FUs MUST NOT be used.
//   The transmission order of single NAL unit packets MUST comply with the NAL unit decoding order.
// 6.3. Non-Interleaved Mode (This mode SHOULD be supported)
//   packetization-mode media type parameter is equal to 1.
//   Only single NAL unit packets, STAP - As, and FU - As MAY be used in this mode.
//   STAP-Bs, MTAPs, and FU-Bs MUST NOT be used.
//   The transmission order of NAL units MUST comply with the NAL unit decoding order
// 6.4. Interleaved Mode
//   packetization-mode media type parameter is equal to 2.
//   STAP-Bs, MTAPs, FU-As, and FU-Bs MAY be used.
//   STAP-As and single NAL unit packets MUST NOT be used.
//   The transmission order of packets and NAL units is constrained as specified in Section 5.5.
//
// 5.1. RTP Header Usage (p10)
// The RTP timestamp is set to the sampling timestamp of the content. A 90 kHz clock rate MUST be used.

package payload

import (
	"errors"
	"github.com/services-go/librtp/rtp"
)

const (
	KHZ = 90
	FU_START = 0x80
	FU_END = 0x40

	N_FU_HEADER = 2
)

type RtpPackH264 struct {
	pkt     rtp.RtpPacket
	handler RtpPayload
	cbparam interface{}
	size    int
}

// create RTP packer
// @param[in] size maximum RTP packet payload size(don't include RTP header)
// @param[in] payload RTP header PT filed (see more about rtp-profile.h)
// @param[in] seq RTP header sequence number filed
// @param[in] ssrc RTP header SSRC filed
// @param[in] handler user-defined callback
// @param[in] cbparam user-defined parameter
// @return RTP packer
func (p RtpPackH264) Create(size int, payload uint8, seq uint16, ssrc uint32, handler RtpPayload, cbparam interface{}) RtpPayloadPacker {
	packer := &RtpPackH264{
		handler: handler,
		size:    size,
	}
	packer.pkt.Header.Version = rtp.RtpVersion
	packer.pkt.Header.PayloadType = payload
	packer.pkt.Header.SequenceNumber = seq
	packer.pkt.Header.SSRC = ssrc
	return packer
}

// destroy RTP Packer
func (p *RtpPackH264) Destroy() {

}

func (p *RtpPackH264) GetInfo() (seq uint16, timestamp uint32) {
	return p.pkt.Header.SequenceNumber, p.pkt.Header.Timestamp
}

// PS/H.264 Elementary Stream to RTP Packet
// @param[in] packer
// @param[in] data stream data
// @param[in] bytes stream length in bytes
// @param[in] time stream UTC time
// @return 0-ok, ENOMEM-alloc failed, <0-failed
func (p *RtpPackH264) Input(data []byte, bytes int, timestamp uint32) error {
	p.pkt.Header.Timestamp = timestamp ﻿//(uint32_t)time * KHz; // ms -> 90KHZ
	var err error
	var p1, p2 []byte
	for p1 = h264NaluFind(data, len(data)); len(p1) > 0 && err == nil; p1 = p2 {
		naluSize := 0
		// ﻿filter H.264 start code(0x00000001)
		p2 = h264NaluFind(p1[1:], len(p1) - 1)
		naluSize = len(p1) - len(p2)

		// ﻿filter suffix '00' bytes
		if len(p2) > 0 {
			naluSize--
		}
		for p1[naluSize-1] == 0 {
			naluSize--
		}

		if naluSize + rtp.RtpFixedHeader <= p.size {
			err = p.rtpH264PackNalu(p1, naluSize)
		} else {
			err = p.rtpH264PackFuA(p1, naluSize)
		}

	}


	return nil
}

func h264NaluFind(data []byte, bytes int) ([]byte) {
	i := 0
	for i += 2; i + 1 < bytes; i++ {
		if data[i] == 0x01 && data[i-1] == 0x00 && data[i-2] == 0x00 {
			return data[i+1:]
		}
	}
	return data[bytes:]
}

func (p *RtpPackH264) rtpH264PackNalu(nalu []byte, bytes int) error {
	p.pkt.Payload = nalu
	p.pkt.PayloadLen = bytes
	n := rtp.RtpFixedHeader + p.pkt.PayloadLen
	rtpb := p.handler.Alloc(p.cbparam, n)
	if rtpb == nil {
		return errors.New("alloc rtp buffer failed.")
	}

	p.pkt.Header.Marker = 0
	if nalu[0]&0x1f <= 5 {
		p.pkt.Header.Marker = 1
	}

	var err error
	n, err = rtp.RtpPacketSerialize(&p.pkt, rtpb, n)
	if err != nil {
		return err
	}
	if n != rtp.RtpFixedHeader+p.pkt.PayloadLen {
		return errors.New("rtp packet serailize failed.")
	}

	p.pkt.Header.SequenceNumber++
	p.handler.Packet(p.cbparam, rtpb, n, p.pkt.Header.Timestamp, 0)
	p.handler.Free(p.cbparam, rtpb)
	return nil
}

func (p *RtpPackH264) rtpH264PackFuA(nalu []byte, bytes int) error {
	﻿// RFC6184 5.3. NAL Unit Header Usage: Table 2 (p15)
	// RFC6184 5.8. Fragmentation Units (FUs) (p29)
	fuIndicator := (nalu[0] & 0xE0) | 28 // FU-A
	fuHeader := (nalu[0] & 0x1F)

	nalu = nalu[1:]
	bytes -= 1
	if bytes <= 0 {
		return errors.New("h264 nalu no bytes.")
	}

	// FU-A start
	var n int
	var err error
	for fuHeader |= FU_START; bytes > 0; p.pkt.Header.SequenceNumber++ {
		p.pkt.PayloadLen = p.size - rtp.RtpFixedHeader - N_FU_HEADER
		if bytes + rtp.RtpFixedHeader <= p.size - N_FU_HEADER {
			if (fuHeader & FU_START) != 0 {
				return errors.New("fuHeader & 0x80 not equal 0.")
			}
			fuHeader = FU_END | (fuHeader & 0x1F)  // FU-U end
			p.pkt.PayloadLen = bytes
		}

		p.pkt.Payload = nalu
		n = rtp.RtpFixedHeader + N_FU_HEADER + p.pkt.PayloadLen
		rtpb := p.handler.Alloc(p.cbparam, n)
		if rtpb == nil {
			return errors.New("alloc rtpb failed.")
		}

		// set marker flag
		p.pkt.Header.Marker = 0
		if FU_END & fuHeader > 0 {
			p.pkt.Header.Marker = 1
		}

		n, err = rtp.RtpPacketSerializeHeader(&p.pkt, rtpb, n)
		if err != nil {
			return err
		}
		if n != rtp.RtpFixedHeader {
			return errors.New("rtp packet serialize failed.")
		}

		// ﻿fu_indicator + fu_header
		rtpb[n] = fuIndicator
		rtpb[1] = fuHeader
		copy(rtpb[n + N_FU_HEADER:], p.pkt.Payload[:p.pkt.PayloadLen])
		p.handler.Packet(p.cbparam, rtpb, n + N_FU_HEADER + p.pkt.PayloadLen, p.pkt.Header.Timestamp, 0)
		p.handler.Free(p.cbparam, rtpb)
		bytes -= p.pkt.PayloadLen
		nalu = nalu[p.pkt.PayloadLen:]
		fuHeader &= 0x1F // clear flags
	}
	return nil
}
