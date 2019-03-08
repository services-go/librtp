package payload

type RtpPayload interface {
	Alloc(param interface{}, bytes int) []byte
	Free(param interface{}, packet []byte)
	Packet(param interface{}, packet []byte, bytes int, timestamp uint32, flags int)
}

type RtpPayloadPacker interface {
	// create RTP packer
	// @param[in] size maximum RTP packet payload size(don't include RTP header)
	// @param[in] payload RTP header PT filed (see more about rtp-profile.h)
	// @param[in] seq RTP header sequence number filed
	// @param[in] ssrc RTP header SSRC filed
	// @param[in] handler user-defined callback
	// @param[in] cbparam user-defined parameter
	// @return RTP packer
	Create(size int, payload uint8, seq uint16, ssrc uint32, handler RtpPayload, cbparam interface{}) RtpPayloadPacker

	// destroy RTP Packer
	Destroy()

	GetInfo() (seq uint16, timestamp uint32)

	// PS/H.264 Elementary Stream to RTP Packet
	// @param[in] packer
	// @param[in] data stream data
	// @param[in] bytes stream length in bytes
	// @param[in] time stream UTC time
	// @return 0-ok, ENOMEM-alloc failed, <0-failed
	Input(data []byte, bytes int, timestamp uint32) error
}

type RtpPayloadUnpacker interface {
	Create(handler RtpPayload, param interface{}) RtpPayloadUnpacker

	Destroy()

	// RTP packet to PS/H.264 Elementary Stream
	// @param[in] decoder RTP packet unpackers
	// @param[in] packet RTP packet
	// @param[in] bytes RTP packet length in bytes
	// @param[in] time stream UTC time
	// @return 1-packet handled, 0-packet discard, <0-failed
	Input(packet []byte, bytes int) error
}
