package payload

type RtpPayloadHelper struct {
	handler   RtpPayload
	cbparam   interface{}
	lost      int    // wait for next frame
	flags     int    // lost packet
	seq       uint16 // rtp seq
	timestamp uint32
	ptr       *uint8
	size      int
	capacity  int
}

func (h *RtpPayloadHelper) Destroy() {
	if h.ptr != nil {
		h.ptr = nil
	}
}
