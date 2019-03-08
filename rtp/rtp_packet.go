package rtp

import "github.com/pkg/errors"

// RFC3550 RTP: A Transport Protocol for Real-Time Applications
// 5.1 RTP Fixed Header Fields (p12)
/*
 0               1               2               3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|V=2|P|X|   CC  |M|     PT      |      sequence number          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                           timestamp                           |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                synchronization source (SSRC) identifier       |
+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
|                 contributing source (CSRC) identifiers        |
|                               ....                            |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

func RtpPacketDeserialize(pkt *RtpPacket, data []byte, bytes int) error {
	// RFC3550 5.1 RTP Fixed Header Fields(p12)
	if bytes < RtpFixedHeader {
		return errors.New("rtp header need 12 bytes.")
	}

	ptr := data
	// pkt header
	v := RtpReadUint32(ptr)
	pkt.Header.Version = RTP_V(v)
	pkt.Header.Padding = RTP_P(v)
	pkt.Header.Extension = RTP_X(v)
	pkt.Header.CSRC = RTP_CC(v)
	pkt.Header.Marker = RTP_M(v)
	pkt.Header.PayloadType = RTP_PT(v)
	pkt.Header.SequenceNumber = RTP_SEQ(v)
	pkt.Header.Timestamp = RtpReadUint32(ptr[4:])
	pkt.Header.SSRC = RtpReadUint32(ptr[8:])

	if pkt.Header.Version != RtpVersion {
		return errors.New("rtp version error.")
	}

	headerlen := RtpFixedHeader + int(pkt.Header.CSRC*4)
	var ext int
	if pkt.Header.Extension > 0 {
		ext += 4
	}
	if pkt.Header.Padding > 0 {
		ext += 1
	}
	if bytes < headerlen+ext {
		return errors.New("no enough bytes.")
	}

	// pkt contributing source
	pkt.CSRC = make([]uint32, pkt.Header.CSRC)
	for i := 0; i < int(pkt.Header.CSRC); i++ {
		pkt.CSRC[i] = RtpReadUint32(ptr[12+i*4:])
	}

	pkt.Payload = ptr[headerlen:]
	pkt.PayloadLen = bytes - headerlen
	// pkt header extension
	if pkt.Header.Extension == 1 {
		rtpext := ptr[headerlen:]
		if pkt.PayloadLen < 4 {
			return errors.New("payload len error.")
		}

		pkt.Extension = rtpext[4:]
		pkt.Reserved = RtpReadUint16(rtpext)
		pkt.Extlen = RtpReadUint16(rtpext[2:]) * 4
		if int(pkt.Extlen+4) > pkt.PayloadLen {
			return errors.New("playload len error2.")
		}
		pkt.Payload = rtpext[pkt.Extlen+4:]
		pkt.PayloadLen -= int(pkt.Extlen + 4)
	}

	// padding
	if pkt.Header.Padding == 1 {
		padding := ptr[bytes-1]
		if pkt.PayloadLen < int(padding) {
			return errors.New("payload len error3.")
		}
		pkt.PayloadLen -= int(padding)
	}
	return nil
}
