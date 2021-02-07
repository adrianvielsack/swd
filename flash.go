package swd

import (
	"encoding/binary"
	"errors"
	"log"
	"time"
)

type Flash struct {
	SWD      *SWD
	writable bool
}

const (
	FLASH_BASE    uint32 = 0x08000000
	FLASH_ACR     uint32 = 0x40022000
	FLASH_KEYR    uint32 = 0x40022004
	FLASH_OPTKEYR uint32 = 0x40022008
	FLASH_SR      uint32 = 0x4002200C

	FLASH_SR_BSY      = (1 << 0)
	FLASH_SR_ERLYBSY  = (1 << 1)
	FLASH_SR_PGERR    = (1 << 2)
	FLASH_SR_WRPRTERR = (1 << 4)
	FLASH_SR_EOP      = (1 << 5)

	FLASH_CR       uint32 = 0x40022010
	FLASH_CR_EOPIE        = 1 << 12

	FLASH_CR_ERRIE  = 1 << 10
	FLASH_CR_OPTWRE = 1 << 9

	FLASH_CR_LOCK  = 1 << 7
	FLASH_CR_STRT  = 1 << 6
	FLASH_CR_OPTER = 1 << 5
	FLASH_CR_OPTPG = 1 << 4
	FLASH_CR_MER   = 1 << 2
	FLASH_CR_PER   = 1 << 1
	FLASH_CR_PG    = 1 << 0

	FLASH_AR   uint32 = 0x40022014
	FLASH_OBR  uint32 = 0x4002201C
	FLASH_WRPR uint32 = 0x40022020

	FLASH_KEY1 = 0x45670123
	FLASH_KEY2 = 0xCDEF89AB
)

func (f *Flash) ReadAddress(addr uint32) ([]byte, error) {
	data, err := f.SWD.ReadRegister(FLASH_BASE + addr)

	if err != nil {
		log.Println("Could not read Flash at addr ", addr)
		return nil, err
	}

	databytes := make([]byte, 4)

	binary.LittleEndian.PutUint32(databytes, data)

	return databytes, nil
}

func NewFlash(SWD *SWD) *Flash {
	return &Flash{
		SWD:      SWD,
		writable: false,
	}
}

func (f *Flash) Read(addr uint32, size uint32) ([]byte, error) {
	size = ((size + 3) / 4) * 4
	data := make([]byte, size)
	for n := uint32(0); n < size; n += 4 {
		d, err := f.SWD.ReadRegister(FLASH_BASE + addr + n)
		if err != nil {
			return nil, err
		}

		binary.LittleEndian.PutUint32(data[n:n+4], d)
	}
	return data, nil
}

func (f *Flash) Writable() error {
	err := f.SWD.WriteRegister(FLASH_KEYR, FLASH_KEY1)
	if err != nil {
		return err
	}
	err = f.SWD.WriteRegister(FLASH_KEYR, FLASH_KEY2)
	if err == nil {
		f.writable = true
		return nil
	}
	return err
}

func (f *Flash) WriteAddress(addr uint32, data []byte) error {
	if !f.writable {
		return errors.New("You musst call the Writable function first")

	}

	if err := f.SWD.WriteRegister(FLASH_CR, FLASH_CR_PG); err != nil {
		return err
	}

	var toWrite uint32
	for n := uint32(0); n < uint32(len(data)); n += 2 {
		toWrite = uint32(binary.LittleEndian.Uint16(data[n : n+2]))
		err := f.SWD.WriteRegisterHalf(FLASH_BASE+n+addr, toWrite<<((n%4)*8))

		if err != nil {
			return err
		}

		busy, _ := f.isBusy()
		for busy {
			busy, err = f.isBusy()
		}
	}

	return nil
}

func (f *Flash) isBusy() (bool, error) {
	val, err := f.SWD.ReadRegister(FLASH_SR)
	if err != nil {
		return false, err
	}
	return val&FLASH_SR_BSY > 0, nil
}

func (f *Flash) EraseAll() error {
	err := f.SWD.WriteRegister(FLASH_CR, FLASH_CR_MER)
	if err != nil {
		return err
	}
	err = f.SWD.WriteRegister(FLASH_CR, FLASH_CR_MER|FLASH_CR_STRT)
	if err != nil {
		return err
	}

	for n := 0; n < 10; n++ {
		busy, err := f.isBusy()
		if err != nil {
			return err
		}
		if !busy {
			return nil
		}
		time.Sleep(time.Millisecond * 100)
	}
	return errors.New("Timeout while waiting for flash to be erased!")
}
