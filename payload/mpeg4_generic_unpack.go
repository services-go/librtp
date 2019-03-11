// RFC3640 RTP Payload Format for Transport of MPEG-4 Elementary Streams
package payload

import (
	"errors"
	"github.com/services-go/librtp/rtp"
)

type RtpUnpackMpeg4Generic struct {
	RtpPayloadHelper
}

func (up *RtpUnpackMpeg4Generic) Init(h RtpPayload, cbparam interface{}) {
	up.handler = h
	up.cbparam = cbparam
	up.flags = -1
}

func (up *RtpUnpackMpeg4Generic) Input(packet []byte, bytes int) (int, error) {
	var pkt rtp.RtpPacket
	err := rtp.RtpPacketDeserialize(&pkt, packet, bytes)
	if err != nil {
		return -1, err
	}
	if pkt.PayloadLen < 4 {
		return -1, errors.New("rtp payload len < 1.")
	}

	up.RtpPayloadCheck(&pkt)
	if up.lost > 0 {
		return 0, nil
	}

	// save payload
	ptr := pkt.Payload
	// AU-headers-length
	auHeaderLen := int(ptr[0])<<8 + int(ptr[1])
	auHeaderLen = (auHeaderLen + 7) / 8 // bit -> byte

	payloadLen := len(ptr)
	if int(auHeaderLen) > payloadLen || auHeaderLen < 2 {
		up.size = 0
		up.lost = 1
		up.flags |= RTP_PAYLOAD_FLAG_PACKET_LOST
		return -1, errors.New("invalid packet.")
	}

	// 3.3.6. High Bit-rate AAC
	// SDP fmtp: sizeLength=13; indexLength=3; indexDeltaLength=3;
	auSize := 2 // only AU-size
	auNumbers := int(auHeaderLen) / auSize
	if int(auHeaderLen)%auSize != 0 {
		return -1, errors.New("au size error.")
	}

	ptr = ptr[2:]            // skip AU headers length section 2-bytes
	pau := ptr[auHeaderLen:] // point to Access Unit

	var size int
	for i := 0; i < auNumbers; i++ {
		size = (int(ptr[0]) << 8) | int(ptr[1]&0xF8)
		size = size >> 3 // bit -> byte
		if size > len(pau) {
			up.size = 0
			up.lost = 1
			up.flags |= RTP_PAYLOAD_FLAG_PACKET_LOST
			return -1, errors.New("rtp payload packet lost.")
		}

		// TODO: add ADTS/ASC ???
		pkt.Payload = pau
		pkt.PayloadLen = size
		up.RtpPayloadWrite(&pkt)

		ptr = ptr[auSize:]
		pau = pau[size:]
		if auNumbers > 1 || pkt.Header.Marker > 0 {
			up.RtpPayloadOnFrame()
		}
	}

	if up.lost > 0 {
		return 0, nil
	}
	return 1, nil
}
