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

import "github.com/services-go/librtp/rtp"

type RtpPackH264 struct {
	Pkt     rtp.RtpPacket
	Handler *RtpPayload
	CbParam interface{}
	Size    int
}

// create RTP packer
// @param[in] size maximum RTP packet payload size(don't include RTP header)
// @param[in] payload RTP header PT filed (see more about rtp-profile.h)
// @param[in] seq RTP header sequence number filed
// @param[in] ssrc RTP header SSRC filed
// @param[in] handler user-defined callback
// @param[in] cbparam user-defined parameter
// @return RTP packer
func (p *RtpPackH264) Create(size int, payload uint8, seq uint16, ssrc uint32, handler RtpPayload, cbparam interface{}) {

}

// destroy RTP Packer
func (p *RtpPackH264) Destroy(packer interface{}) {

}

func (p *RtpPackH264) GetInfo(packer interface{}) (seq uint16, timestamp uint32) {
	return 0, 0
}

// PS/H.264 Elementary Stream to RTP Packet
// @param[in] packer
// @param[in] data stream data
// @param[in] bytes stream length in bytes
// @param[in] time stream UTC time
// @return 0-ok, ENOMEM-alloc failed, <0-failed
func (p *RtpPackH264) Input(packer interface{}, data interface{}, bytes int, timestamp uint32) {

}
