package payload

type RtpUnpackH264 struct {
	handler  *RtpPayload
	cbparam  interface{}
	seq      uint16
	ptr      *byte
	size     int
	capacity int
	flags    int
}

func (up *RtpUnpackH264) Create(handler RtpPayload, param interface{}) {

}

func (up *RtpUnpackH264) Destroy(packer interface{}) {

}

func (up *RtpUnpackH264) Input(packer interface{}) {

}
