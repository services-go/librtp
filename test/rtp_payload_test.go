package test

import (
	"fmt"
	"github.com/services-go/librtp/payload"
	"log"
	"os"
	"strings"
	"testing"
)

type RtpPayloadTest struct {
	payload  int
	encoding string

	frtp    *os.File
	fsource *os.File

	frtp2    *os.File
	fsource2 *os.File

	payloadDe *payload.RtpPayloadDelegate

	size   int
	packet []byte
}

type payloadPacker struct{}

type payloadUnpacker struct{}

func (p *payloadPacker) Alloc(param interface{}, bytes int) []byte {
	buf := make([]byte, bytes+4) // 0x00 0x00 0x00 0x01
	return buf
}

func (p *payloadPacker) Free(param interface{}, packet []byte) {
}

func (p *payloadPacker) Handle(param interface{}, packet []byte, bytes int, timestamp uint32, flags int) {
	size := make([]byte, 2)
	size[0] = byte(bytes >> 8)
	size[1] = byte(bytes)

	ctx, _ := param.(*RtpPayloadTest)
	fmt.Printf("22 frtp2 = %p\n", ctx.frtp2)
	ctx.frtp2.Write(size)
	ctx.frtp2.Write(packet)
}

func (up *payloadUnpacker) Alloc(param interface{}, bytes int) []byte {
	buf := make([]byte, bytes+4) // 0x00 0x00 0x00 0x01
	return buf
}

func (up *payloadUnpacker) Free(param interface{}, packet []byte) {
}

func (up *payloadUnpacker) Handle(param interface{}, packet []byte, bytes int, timestamp uint32, flags int) {
	startcode := []byte{0x00, 0x00, 0x00, 0x01}
	buffer := make([]byte, bytes+12)

	ctx, _ := param.(*RtpPayloadTest)
	size := 0
	if strings.Compare("H264", ctx.encoding) == 0 || strings.Compare("H265", ctx.encoding) == 0 {
		copy(buffer, startcode)
		size += len(startcode)
	} else if strings.Compare("mpeg4-generic", ctx.encoding) == 0 {
		len := bytes + 7
		var profile byte = 2
		var samplingFrenquecyIndex byte = 4
		var channelConfiguration byte = 2
		buffer[0] = 0xFF // 12-syncword
		buffer[1] = 0xF0 | (0 << 3) | (0x00 << 2) | 0x01
		buffer[2] = ((profile - 1) << 6) | ((samplingFrenquecyIndex & 0x0F) << 2) | ((channelConfiguration >> 2) & 0x01)
		buffer[3] = ((channelConfiguration & 0x03) << 6) | byte(len>>11)&0x03
		buffer[4] = byte(len >> 3)
		buffer[5] = (byte(len)&0x07)<<5 | 0x1F
		buffer[6] = 0xFC | byte((len/1024)&0x03)
		size = 7
	}
	copy(buffer[size:], packet)
	size += bytes

	// TODO:
	// check media file
	_, err := ctx.fsource2.Write(buffer[:size])
	if err != nil {
		log.Println("write error:", err)
	}
	ctx.payloadDe.RtpPayloadPackerInput(buffer, size, timestamp)
}

func rtpPayloadTest(pt int, encoding string, seq uint16, ssrc uint32, rtpfile string, sourcefile string) {
	var ctx RtpPayloadTest
	ctx.payload = pt
	ctx.encoding = encoding

	var err error
	ctx.frtp, err = os.OpenFile(rtpfile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err.Error() + ":" + rtpfile)
		return
	}
	defer ctx.frtp.Close()

	ctx.fsource, err = os.OpenFile(sourcefile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err.Error() + ":" + sourcefile)
		return
	}
	defer ctx.fsource.Close()

	ctx.frtp2, err = os.OpenFile("out.rtp", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err)
		return

	}
	defer ctx.frtp2.Close()
	fmt.Printf("11 frtp2 = %p\n", ctx.frtp2)

	ctx.fsource2, err = os.OpenFile("out.media", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer ctx.fsource2.Close()

	fmt.Printf("11 ctx fsource2 = %p\n", ctx.fsource2)

	var p payloadPacker
	var up payloadUnpacker
	ctx.payloadDe, err = payload.RtpPayloadCreate(pt, encoding, seq, ssrc, 1456, &p, &up, &ctx)
	if err != nil {
		log.Println(err)
		return
	}

	ctx.packet = make([]byte, 1024*1024)
	for {
		sz := make([]byte, 2)
		n, err := ctx.frtp.Read(sz)
		if err != nil {
			log.Println(err.Error())
			break
		}

		ctx.size = int(sz[0])<<8 | int(sz[1])

		n, err = ctx.frtp.Read(ctx.packet[:ctx.size])
		if err != nil || n != ctx.size {
			log.Println(err)
			break
		}
		_, err = ctx.payloadDe.RtpPayloadUnpackerInput(ctx.packet, ctx.size)
		if err != nil {
			log.Println(err)
		}
	}
	ctx.payloadDe.RtpPayloadPackerDestroy()
	ctx.payloadDe.RtpPayloadUnpackerDestroy()
	log.Println("write file over..")
}

func TestRtpPayload(t *testing.T) {
	rtpPayloadTest(96, "H264", 12686, 1957754144, "./girls704x576_base.rtp", "./girls704x576_base.h264")
}
