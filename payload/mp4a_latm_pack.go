package payload

import "github.com/services-go/librtp/rtp"

type RtpPackMp4aLatm struct {
	handler *RtpPayload
	pkt     rtp.RtpPacket
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
func (p RtpPackMp4aLatm) Create(size int, payload uint8, seq uint16, ssrc uint32, handler RtpPayload, cbparam interface{}) RtpPayloadPacker {

}

// destroy RTP Packer
func (p *RtpPackMp4aLatm) Destroy() {

}

func (p *RtpPackMp4aLatm) GetInfo() (seq uint16, timestamp uint32) {
	return 0, 0
}

// PS/H.264 Elementary Stream to RTP Packet
// @param[in] packer
// @param[in] data stream data
// @param[in] bytes stream length in bytes
// @param[in] time stream UTC time
// @return 0-ok, ENOMEM-alloc failed, <0-failed
func (p *RtpPackMp4aLatm) Input(data []byte, bytes int, timestamp uint32) {

}
