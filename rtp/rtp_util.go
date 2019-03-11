package rtp

func RtpReadUint16(ptr []byte) uint16 {
	return uint16(ptr[0])<<8 | uint16(ptr[1])
}

func RtpReadUint32(ptr []byte) uint32 {
	return uint32(ptr[0])<<24 | uint32(ptr[1])<<16 | uint32(ptr[2])<<8 | uint32(ptr[3])
}

func RtpWriteUint16(ptr []byte, val uint16) {
	ptr[0] = byte(val >> 8)
	ptr[1] = byte(val)
}
func RtpWriteUint32(ptr []byte, val uint32) {
	ptr[0] = byte(val >> 24)
	ptr[1] = byte(val >> 16)
	ptr[2] = byte(val >> 8)
	ptr[3] = byte(val)
}

func WriteRtpHeader(ptr []byte, h *RtpHeader) {
	ptr[0] = (h.Version << RtpHeader_VersionShift) | (h.Padding << RtpHeader_PaddingShift) |
		(h.Extension << RtpHeader_ExtensionShift) | h.CSRC

	ptr[1] = (h.Marker << 7) | h.PayloadType
	ptr[2] = byte(h.SequenceNumber >> 8)
	ptr[3] = byte(h.SequenceNumber & 0xff)
	RtpWriteUint32(ptr[4:], h.Timestamp)
	RtpWriteUint32(ptr[8:], h.SSRC)
}

//func WriteRtcpHeader(ptr []byte, h *RtcpHeader) {
////
////}

/*
static inline void nbo_write_rtcp_header(uint8_t *ptr, const rtcp_header_t *header)
{
	ptr[0] = (uint8_t)((header->v << 6) | (header->p << 5) | header->rc);
	ptr[1] = (uint8_t)(header->pt);
	ptr[2] = (uint8_t)(header->length >> 8);
	ptr[3] = (uint8_t)(header->length & 0xFF);
}
*/
