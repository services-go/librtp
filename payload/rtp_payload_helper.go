package payload

import (
	"errors"
	"github.com/services-go/librtp/rtp"
)

type RtpPayloadHelper struct {
	handler   RtpPayload
	cbparam   interface{}
	lost      int    // wait for next frame
	flags     int    // lost packet
	seq       uint16 // rtp seq
	timestamp uint32
	ptr       []byte
	size      int
	capacity  int
}

func (h *RtpPayloadHelper) Destroy() {
	if h.ptr != nil {
		h.ptr = nil
	}
}

func (h *RtpPayloadHelper) RtpPayloadCheck(pkt *rtp.RtpPacket) {
	// first packet only
	if h.flags == -1 {
		h.flags = 0
		h.seq = pkt.Header.SequenceNumber - 1  // disable packet lost
		h.timestamp = pkt.Header.Timestamp + 1 // flag for new frame
	}

	// check sequence number
	if pkt.Header.SequenceNumber != h.seq+1 {
		h.size = 0
		h.lost = 1
		h.flags |= RTP_PAYLOAD_FLAG_PACKET_LOST
		h.timestamp = pkt.Header.Timestamp
	}
	h.seq = pkt.Header.SequenceNumber

	// check timestamp
	if pkt.Header.Timestamp != h.timestamp {
		h.RtpPayloadOnFrame()
	}
	h.timestamp = pkt.Header.Timestamp
}

func (h *RtpPayloadHelper) RtpPayloadWrite(pkt *rtp.RtpPacket) error {
	if h.size+pkt.PayloadLen > h.capacity {
		size := h.size + pkt.PayloadLen + 8000
		ptr := make([]byte, size)
		if h.ptr != nil {
			copy(ptr, h.ptr)
		}
		h.ptr = ptr
		h.capacity = size
	}

	if h.capacity < h.size+pkt.PayloadLen {
		return errors.New("payload helper capacity error.")
	}
	copy(h.ptr[h.size:], pkt.Payload[:pkt.PayloadLen])
	h.size += pkt.PayloadLen
	return nil
}

func (h *RtpPayloadHelper) RtpPayloadOnFrame() {
	if h.size > 0 {
		// previous packet done
		h.handler.Handle(h.cbparam, h.ptr, h.size, h.timestamp, h.flags)
		h.flags = 0 // clear packet lost flag
	}

	// new frame start
	h.lost = 0
	h.size = 0
}
