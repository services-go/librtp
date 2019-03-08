package rtp

const (
	RtpVersion     = 2  // RTP version field must equal 2 (p66)
	RtpFixedHeader = 12 // the header fixed 12 byte
)

const (
	RtpHeader_HeaderLength    = 3 // Fixed header len 12 bytes
	RtpHeader_VersionShift    = 6
	RtpHeader_VersionMask     = 0x3
	RtpHeader_PaddingShift    = 5
	RtpHeader_PaddingMask     = 0x1
	RtpHeader_ExtensionShift  = 4
	RtpHeader_ExtensionMask   = 0x1
	RtpHeader_CCMask          = 0xF
	RtpHeader_MarkerShift     = 7
	RtpHeader_MarkerMask      = 0x1
	RtpHeader_PtMask          = 0x7F
	RtpHeader_SeqNumOffset    = 2
	RtpHeader_SeqNumLength    = 2
	RtpHeader_TimestampOffset = 4
	RtpHeader_TimestampLength = 4
	RtpHeader_SsrcOffset      = 8
	RtpHeader_SsrcLength      = 4
	RtpHeader_CsrcOffset      = 12
	RtpHeader_CsrcLength      = 4
)

type RtpHeader struct {
	Version        byte   // protocol version
	Padding        byte   // padding flag
	Extension      byte   // header extension flag
	CSRC           byte   // CSRC count
	Marker         byte   // marker bit
	PayloadType    byte   // payload type
	SequenceNumber uint16 // sequence number
	Timestamp      uint32 // timestamp
	SSRC           uint32 // synchronization source
}

type RtpPacket struct {
	Header     RtpHeader
	CSRC       []uint32
	Extension  []byte // extension(valid only if rtp.x = 1)
	Extlen     uint16 // extension length in bytes
	Reserved   uint16 // extension reserved
	Payload    []byte // payload
	PayloadLen int    //payload length in bytes
}

func RTP_V(v uint32) byte {
	return byte((v >> 30) & 0x03)
}

func RTP_P(v uint32) byte {
	return byte((v >> 29) & 0x01)
}

func RTP_X(v uint32) byte {
	return byte((v >> 28) & 0x01)
}

func RTP_CC(v uint32) byte {
	return byte((v >> 24) & 0x0F)
}

func RTP_M(v uint32) byte {
	return byte((v >> 23) & 0x01)
}

func RTP_PT(v uint32) byte {
	return byte((v >> 16) & 0x7F)
}

func RTP_SEQ(v uint32) uint16 {
	return uint16((v >> 00) & 0xFFFF)
}
