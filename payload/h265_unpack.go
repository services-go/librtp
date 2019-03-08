package payload

type RtpUnpackH265 struct {
	handler  *RtpPayload
	cbparam  interface{}
	seq      uint16
	ptr      *byte
	size     int
	capacity int
	flags    int
}

func (up *RtpUnpackH265) Create(handler RtpPayload, param interface{}) {

}

func (up *RtpUnpackH265) Destroy() {

}

func (up *RtpUnpackH265) Input(packet []byte, bytes int) error {

}
