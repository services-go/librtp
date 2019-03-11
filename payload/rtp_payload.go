package payload

import (
	"errors"
	"github.com/services-go/librtp/rtp"
	"strconv"
)

const (
	RTP_PAYLOAD_FLAG_PACKET_LOST = 1
)

type RtpPayload interface {
	Alloc(param interface{}, bytes int) []byte
	Free(param interface{}, packet []byte)
	Handle(param interface{}, packet []byte, bytes int, timestamp uint32, flags int)
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
	Init(size int, payload uint8, seq uint16, ssrc uint32, handler RtpPayload, cbparam interface{})

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
	Init(handler RtpPayload, param interface{})

	Destroy()

	// RTP packet to PS/H.264 Elementary Stream
	// @param[in] decoder RTP packet unpackers
	// @param[in] packet RTP packet
	// @param[in] bytes RTP packet length in bytes
	// @param[in] time stream UTC time
	// @return 1-packet handled, 0-packet discard, <0-failed
	Input(packet []byte, bytes int) (int, error)
}

type RtpPayloadDelegate struct {
	Packer   RtpPayloadPacker
	Unpacker RtpPayloadUnpacker
	//RtpMaxPacketSize int // default 1434 from VLC
}

func RtpPayloadCreate(payload int, name string, seq uint16, ssrc uint32, packsize int,
	packhandler RtpPayload, unpackhandler RtpPayload, cbparam interface{}) (*RtpPayloadDelegate, error) {
	delegate := &RtpPayloadDelegate{}
	err := delegate.rtpPayloadFind(payload, name)
	if err != nil {
		return nil, err
	}
	//size := delegate.RtpPacketGetSize()
	delegate.Packer.Init(packsize, uint8(payload), seq, ssrc, packhandler, cbparam)
	delegate.Unpacker.Init(unpackhandler, cbparam)
	return delegate, nil
}

func (de *RtpPayloadDelegate) RtpPayloadPackerDestroy() {
	de.Packer.Destroy()
}

func (de *RtpPayloadDelegate) RtpPayloadPackerGetInfo() (seq uint16, timestamp uint32) {
	return de.Packer.GetInfo()
}

func (de *RtpPayloadDelegate) RtpPayloadPackerInput(data []byte, bytes int, timestamp uint32) error {
	return de.Packer.Input(data, bytes, timestamp)
}

func (de *RtpPayloadDelegate) RtpPayloadUnpackerDestroy() {
	de.Unpacker.Destroy()
}

func (de *RtpPayloadDelegate) RtpPayloadUnpackerInput(packet []byte, bytes int) (int, error) {
	return de.Unpacker.Input(packet, bytes)
}

//func (de *RtpPayloadDelegate) RtpPacketSetSize(bytes int) {
//	de.RtpMaxPacketSize = bytes
//}
//
//func (de *RtpPayloadDelegate) RtpPacketGetSize() int {
//	return de.RtpMaxPacketSize
//}

func (de *RtpPayloadDelegate) rtpPayloadFind(payload int, encoding string) error {
	if payload < 0 || payload > 127 {
		return errors.New("rtp error payload: " + strconv.Itoa(payload))
	}
	if payload >= 96 && len(encoding) > 0 {
		switch encoding {
		case "H264":
			// H.264 video (MPEG-4 Part 10) (RFC 6184)
			de.Packer = &RtpPackH264{}
			de.Unpacker = &RtpUnpackH264{}
		case "H265":
		case "HEVC":
			// H.265 video (HEVC) (RFC 7798)
			return errors.New("not support h265.")
		case "mpeg4-generic":
		case "AAC":
			/// RFC3640 RTP Payload Format for Transport of MPEG-4 Elementary Streams
			/// 4.1. MIME Type Registration (p27)
			de.Packer = &RtpPackMpeg4Generic{}
			de.Unpacker = &RtpUnpackMpeg4Generic{}
		case "OPUS": // RFC7587 RTP Payload Format for the Opus Speech and Audio Codec
		case "G726-16": // ITU-T G.726 audio 16 kbit/s (RFC 3551)
		case "G726-24": // ITU-T G.726 audio 24 kbit/s (RFC 3551)
		case "G726-32": // ITU-T G.726 audio 32 kbit/s (RFC 3551)
		case "G726-40": // ITU-T G.726 audio 40 kbit/s (RFC 3551)
		case "G7221": // RFC5577 RTP Payload Format for ITU-T Recommendation G.722.1
			de.Packer = &RtpCommPack{}
			de.Unpacker = &RtpCommUnpack{}
		default:
			return errors.New("not support code: " + encoding)
		}
	} else {
		switch payload {
		case rtp.RTP_PAYLOAD_PCMU: // ITU-T G.711 PCM u-Law audio 64 kbit/s (RFC 3551)
		case rtp.RTP_PAYLOAD_PCMA: // ITU-T G.711 PCM A-Law audio 64 kbit/s (RFC 3551)
		case rtp.RTP_PAYLOAD_G722: // ITU-T G.722 audio 64 kbit/s (RFC 3551)
		case rtp.RTP_PAYLOAD_G729: // ITU-T G.729 and G.729a audio 8 kbit/s (RFC 3551)
			de.Packer = &RtpCommPack{}
			de.Unpacker = &RtpCommUnpack{}
		default:
			return errors.New("not support payload: " + strconv.Itoa(payload))
		}
	}
	return nil
}
