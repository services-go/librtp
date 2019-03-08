package payload

type RtpUnpack struct {
	RtpPayloadHelper
}

// 对应 rtp_decode_rfc2250
func (up *RtpUnpack) Input(p interface{}, packet interface{}, bytes int) {

}
