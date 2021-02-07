package lowlevel

import (
	"errors"
	"fmt"
	"github.com/stianeikeland/go-rpio"
	"time"
)

const (
	MAGIC_JTAG_TO_SWD_PART1      = 0x9e //0x79
	MAGIC_JTAG_TO_SWD_PART2      = 0xe7
	SWD_START               byte = 0b00000001
	SWD_APnDP               byte = 0b00000010
	SWD_RnW                 byte = 0b00000100
	SWD_A_2                 byte = 0b00001000
	SWD_A_3                 byte = 0b00010000
	SWD_PARITY              byte = 0b00100000
	SWD_PARK                byte = 0b10000000

	SWD_OK_RESPONSE    byte = 0b001
	SWD_WAIT_RESPONSE  byte = 0b010
	SWD_FAULT_RESPONSE byte = 0b100
)

type SwdRaspberrygpio struct {
	PinSWDIO       rpio.Pin
	PinSWDCLK      rpio.Pin
	halfClockCycle time.Duration
}

func NewSwdRaspberrygpio(PinSWDIO int, PinSWDCLK int, freq int) (*SwdRaspberrygpio, error) {
	swd := &SwdRaspberrygpio{
		PinSWDIO:       rpio.Pin(PinSWDIO),
		PinSWDCLK:      rpio.Pin(PinSWDCLK),
		halfClockCycle: time.Second / time.Duration(freq),
	}

	if err := rpio.Open(); err != nil {
		return nil, err
	}

	swd.PinSWDCLK.Mode(rpio.Output)
	swd.PinSWDCLK.Low()
	swd.PinSWDCLK.PullDown()
	swd.PinSWDCLK.Mode(rpio.Output)
	swd.PinSWDIO.PullUp()
	swd.Initialize()

	return swd, nil
}

func (swd *SwdRaspberrygpio) sendBit(bit byte) {
	swd.PinSWDIO.Output()
	time.Sleep(swd.halfClockCycle)
	swd.PinSWDCLK.Low()
	bit = bit & 0x1
	if bit == 1 {
		swd.PinSWDIO.High()
	} else {
		swd.PinSWDIO.Low()
	}
	time.Sleep(swd.halfClockCycle)
	swd.PinSWDCLK.High()

}

func (swd *SwdRaspberrygpio) readBit() byte {
	ret := byte(0)
	swd.PinSWDIO.Input()

	time.Sleep(swd.halfClockCycle)
	swd.PinSWDCLK.Low()
	if swd.PinSWDIO.Read() == rpio.High {
		ret = 1
	}
	time.Sleep(swd.halfClockCycle)
	swd.PinSWDCLK.High()
	return ret
}

func (swd *SwdRaspberrygpio) Reset() {

	for i := 0; i < 8; i++ {
		swd.sendByte(0xff)
	}

	for i := 0; i < 8; i++ {
		swd.sendByte(0x00)
	}

	for i := 0; i < 8; i++ {
		swd.sendByte(0xff)
	}
}

func (swd *SwdRaspberrygpio) Sync() {

	for i := 0; i < 8; i++ {
		swd.sendByte(0xff)
	}

	for i := 0; i < 8; i++ {
		swd.sendByte(0x00)
	}

}

func (swd *SwdRaspberrygpio) sendArbByte(data byte, len int) {
	if len > 8 {
		len = 8
	}

	for i := 0; i < len; i++ {
		swd.sendBit(data)
		data = data >> 1
	}
}

func (swd *SwdRaspberrygpio) sendByte(data byte) {
	swd.sendArbByte(data, 8)
}

func (swd *SwdRaspberrygpio) sendBytes(data []byte) {
	for _, b := range data {
		swd.sendByte(b)
	}
}

func (swd *SwdRaspberrygpio) readArbBits(len int) byte {
	if len > 8 {
		len = 8
	}
	ret := byte(0)

	for i := 0; i < len; i++ {
		ret |= (swd.readBit() & 0x01) << i
	}

	return ret
}

func (swd *SwdRaspberrygpio) readArbBytes(len int) []byte {
	ret := make([]byte, (len+7)/8)
	for b := 0; b < (len+7)/8; b++ {
		ret[b] = swd.readArbBits(len - (b * 8))
	}
	return ret
}

func (swd *SwdRaspberrygpio) writeArbBytes(data []byte, len int) {

	for b := 0; b < (len+7)/8; b++ {
		swd.sendArbByte(data[b], len-(b*8))
	}
}

func (swd *SwdRaspberrygpio) readByte() byte {
	return swd.readArbBits(8)
}

func (swd *SwdRaspberrygpio) getStartByteParity(val byte) byte {
	return ((val << 4) ^ (val << 3) ^ (val << 2) ^ (val << 1)) & 0x20
}

func (swd *SwdRaspberrygpio) getDataParityBit(data []byte) byte {
	parity := byte(0)
	for b := 0; b < 4; b++ {
		for i := 0; i < 8; i++ {
			parity ^= data[b] >> i
		}
	}
	return parity & 0x1
}

func (swd *SwdRaspberrygpio) SWDRead(ap bool, address byte) (uint32, error) {

	address = (address << 1) & 0x18

	startByte := SWD_START | SWD_RnW | address | SWD_PARK
	if ap {
		startByte |= SWD_APnDP
	}
	startByte |= swd.getStartByteParity(startByte)
	swd.sendByte(startByte)

	swd.readBit() // Trn
	ACK := swd.readArbBits(3)
	if ACK != SWD_OK_RESPONSE {
		return 0, errors.New("Didn't receive SWD Response OK")
	}

	ret := swd.readArbBytes(33)
	swd.sendBit(0)
	swd.sendBit(0)
	if len(ret) != 5 {
		return 0, errors.New(fmt.Sprintf("Receive errro, expected 5 bytes, got %d", len(ret)))
	}

	parity := swd.getDataParityBit(ret)
	if (ret[4] & 0x01) != parity {
		return 0, errors.New(fmt.Sprintf("Parity bit of received data does not match expected"))
	}

	data := uint32(ret[0]) | (uint32(ret[1]) << 8) | (uint32(ret[2]) << 16) | (uint32(ret[3]) << 24)

	return data, nil
}

func (swd *SwdRaspberrygpio) SWDWrite(ap bool, cmd byte, data uint32) error {
	cmd = (cmd << 1) & 0x18
	startByte := SWD_START | cmd | SWD_PARK
	if ap {
		startByte |= SWD_APnDP
	}
	startByte |= swd.getStartByteParity(startByte)
	toSend := make([]byte, 5)
	for i := 0; i < 4; i++ {
		toSend[i] = byte((data >> (i * 8)) & 0xFF)
	}
	toSend[4] = swd.getDataParityBit(toSend)
	swd.sendByte(startByte)

	swd.readBit() // Trn
	ACK := swd.readArbBits(3) & 0x07
	if ACK != SWD_OK_RESPONSE {
		return errors.New("Error writing to SWD, didn't receive SWD_OK")
	}
	swd.readBit() // Trn

	swd.writeArbBytes(toSend, 33)

	//Idle Line State after Write
	swd.sendByte(0)

	return nil
}

func (swd *SwdRaspberrygpio) Initialize() {
	swd.Reset()
	swd.sendByte(MAGIC_JTAG_TO_SWD_PART1)
	swd.sendByte(MAGIC_JTAG_TO_SWD_PART2)
	swd.Sync()

}
