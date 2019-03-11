package payload

import (
	"errors"
	"github.com/services-go/librtp/rtp"
)

type RtpUnpackH264 struct {
	handler  RtpPayload
	cbparam  interface{}
	seq      uint16
	ptr      []byte
	size     int
	capacity int
	flags    int
}

func (up *RtpUnpackH264) Init(handler RtpPayload, param interface{}) {
	up.handler = handler
	up.cbparam = param
	up.flags = -1
}

func (up *RtpUnpackH264) Destroy() {
	if up.ptr != nil {
		up.ptr = nil
	}
}

func (up *RtpUnpackH264) Input(data []byte, bytes int) (int, error) {
	var pkt rtp.RtpPacket
	err := rtp.RtpPacketDeserialize(&pkt, data, bytes)
	if err != nil {
		return -1, err
	}
	if pkt.PayloadLen < 1 {
		return -1, errors.New("payload len < 1.")
	}

	if up.flags == -1 {
		up.flags = 0
		up.seq = pkt.Header.SequenceNumber - 1 // disable packet lost
	}

	if pkt.Header.SequenceNumber != up.seq+1 {
		up.flags = RTP_PAYLOAD_FLAG_PACKET_LOST
		up.size = 0 // discard previous packets
	}
	up.seq = pkt.Header.SequenceNumber

	nal := pkt.Payload[0]
	switch nal & 0x1F {
	case 0: // reserved
	case 31: // reserved
		return 0, nil // packet discard
	case 24: // STAP-A
		return up.rtpH264UnpackStap(pkt.Payload, pkt.PayloadLen, pkt.Header.Timestamp, 0)
	case 25: // STAP-B
		return up.rtpH264UnpackStap(pkt.Payload, pkt.PayloadLen, pkt.Header.Timestamp, 1)
	case 26: // MTAP16
		return up.rtpH264UnpackMtap(pkt.Payload, pkt.PayloadLen, pkt.Header.Timestamp, 2)
	case 27: // MTAP24
		return up.rtpH264UnpackMtap(pkt.Payload, pkt.PayloadLen, pkt.Header.Timestamp, 3)
	case 28: // FU-A
		return up.rtpH264UnpackFu(pkt.Payload, pkt.PayloadLen, pkt.Header.Timestamp, 0)
	case 29: // FU-B
		return up.rtpH264UnpackFu(pkt.Payload, pkt.PayloadLen, pkt.Header.Timestamp, 1)
	default: // 1-23 NAL unit
		up.handler.Handle(up.cbparam, pkt.Payload, pkt.PayloadLen, pkt.Header.Timestamp, up.flags)
		up.flags = 0
		up.size = 0
	}
	return 1, nil
}

// 5.7.1. Single-Time Aggregation Packet (STAP) (p23)
/*
 0               1               2               3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                           RTP Header                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|STAP-B NAL HDR |            DON                |  NALU 1 Size  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| NALU 1 Size   | NALU 1 HDR    |         NALU 1 Data           |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+                               +
:                                                               :
+               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|               | NALU 2 Size                   |   NALU 2 HDR  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                            NALU 2 Data                        |
:                                                               :
|                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                               :    ...OPTIONAL RTP padding    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
func (up *RtpUnpackH264) rtpH264UnpackStap(ptr []byte, bytes int, timestamp uint32, stapb int) (int, error) {
	n := 1
	if stapb != 0 {
		n = 3
	}

	don := 0
	if stapb != 0 {
		don = int(rtp.RtpReadUint16(ptr[1:]))
	}
	ptr = ptr[n:] // STAP-A / STAP-B HDR + DON

	var len int
	for bytes -= n; bytes > 2; bytes -= len + 2 {
		len = int(rtp.RtpReadUint16(ptr))
		if len+2 > bytes {
			up.flags = RTP_PAYLOAD_FLAG_PACKET_LOST
			up.size = 0
			return -1, errors.New("1 payload flag packet lost.")
		}

		if H264_NAL_264V(ptr[2]) <= 0 || H264_NAL_264V(ptr[2]) >= 24 {
			return -1, errors.New("h264 nal error.")
		}

		up.handler.Handle(up.cbparam, ptr[2:], len, timestamp, up.flags)
		up.flags = 0
		up.size = 0

		ptr = ptr[len+2:] // next NALU
		don = (don + 1) % 65536
	}
	return 1, nil
}

// 5.7.2. Multi-Time Aggregation Packets (MTAPs) (p27)
/*
 0               1               2               3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                          RTP Header                           |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|MTAP16 NAL HDR |   decoding order number base  |  NALU 1 Size  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| NALU 1 Size   | NALU 1 DOND   |         NALU 1 TS offset      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| NALU 1 HDR    |                NALU 1 DATA                    |
+-+-+-+-+-+-+-+-+                                               +
:                                                               :
+               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|               | NALU 2 SIZE                   |   NALU 2 DOND |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| NALU 2 TS offset              | NALU 2 HDR    |  NALU 2 DATA  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+               |
:                                                               :
|                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                               :    ...OPTIONAL RTP padding    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

func (up *RtpUnpackH264) rtpH264UnpackMtap(ptr []byte, bytes int, timestamp uint32, n int) (int, error) {
	//donb := rtp.RtpReadUint16(ptr[1:])
	ptr = ptr[3:] // MTAP16/MTAP24 HDR + DONB

	var len int
	//var dond uint16
	var ts uint32
	for bytes -= 3; bytes > 3+n; bytes -= len + 2 {
		len = int(rtp.RtpReadUint16(ptr))
		// 1 - DOND n - TS offset 1 - NALU
		if len+2 > bytes || len < 1+n+1 {
			up.flags = RTP_PAYLOAD_FLAG_PACKET_LOST
			up.size = 0
			return -1, errors.New("2 rtp payload flag packet lost.")
		}

		//dond = uint16(int(uint16(ptr[2])+donb) % 65536)
		ts = uint32(rtp.RtpReadUint16(ptr[3:]))
		if n == 3 {
			ts = (ts << 16) | uint32(ptr[5]) // MTAP24
		}

		// if the NALU-time is larger than or equal to the RTP timestamp of the packet,
		// then the timestamp offset equals (the NALU - time of the NAL unit - the RTP timestamp of the packet).
		// If the NALU - time is smaller than the RTP timestamp of the packet,
		// then the timestamp offset is equal to the NALU - time + (2 ^ 32 - the RTP timestamp of the packet).
		ts += timestamp // wrap 1 << 32

		if H264_NAL_264V(ptr[n+3]) <= 0 || H264_NAL_264V(ptr[n+3]) >= 24 {
			return -1, errors.New("h264 nalu error.")
		}
		up.handler.Handle(up.cbparam, ptr[1+n:], len-1-n, ts, up.flags)
		up.flags = 0
		up.size = 0

		ptr = ptr[len+1+n:]
	}
	return 1, nil
}

// 5.8. Fragmentation Units (FUs) (p29)
/*
 0               1               2               3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|  FU indicator |   FU header   |              DON              |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-|
|                                                               |
|                          FU payload                           |
|                                                               |
|                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                               :   ...OPTIONAL RTP padding     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
func (up *RtpUnpackH264) rtpH264UnpackFu(ptr []byte, bytes int, timestamp uint32, fub int) (int, error) {
	n := 2
	if fub != 0 {
		n = 4
	}
	if bytes < n {
		return -1, errors.New("error unpack bytes.")
	}

	if up.size+bytes-n+1 > up.capacity {
		size := up.size + bytes + 128000 + 1
		p := make([]byte, size)
		if up.ptr != nil {
			copy(p, up.ptr)
		}
		up.ptr = p
		up.capacity = size
	}

	fuheader := ptr[1]
	if FU_START_264V(fuheader) != 0 {
		up.size = 1 // NAL unit type byte
		up.ptr[0] = (ptr[0] & 0xE0) | (fuheader & 0x1F)
		if H264_NAL_264V(up.ptr[0]) <= 0 || H264_NAL_264V(up.ptr[0]) >= 24 {
			return -1, errors.New("h264 nalu error.")
		}
	} else {
		if up.size == 0 {
			up.flags = RTP_PAYLOAD_FLAG_PACKET_LOST
			return -1, errors.New("1 rtp payload flag packet lost.")
		}
	}

	if bytes > n {
		copy(up.ptr[up.size:], ptr[n:])
		up.size += bytes - n
	}

	if FU_END_264V(fuheader) != 0 {
		up.handler.Handle(up.cbparam, up.ptr, up.size, timestamp, up.flags)
		up.flags = 0
		up.size = 0 // reset
	}

	return 1, nil
}

func H264_NAL_264V(v byte) byte {
	return v & 0x1F
}

func FU_START_264V(v byte) byte {
	return v & 0x80
}

func FU_END_264V(v byte) byte {
	return v & 0x40
}

func FU_NAL_264V(v byte) byte {
	return v & 0x1F
}
