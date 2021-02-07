package swd

import "log"

type STM32 struct {
	SWD       *SWD
	CoreDebug *CoreDebug
	Flash     *Flash
}

func NewSTM32(swd *SWD) *STM32 {
	return &STM32{
		SWD:       swd,
		CoreDebug: &CoreDebug{swd},
		Flash:     NewFlash(swd),
	}
}

func (stm *STM32) Initialize() error {
	stmId, err := stm.SWD.ReadID()
	if err != nil {
		log.Fatalln("Could not read Chip ID")
		log.Printf("Chip ID: 0x%08X ?", stmId)
		return err
	}
	if err := stm.SWD.PowerOnRequest(); err != nil {
		log.Fatalln("Could not request Power On: " + err.Error())
		return err
	}
	if err := stm.SWD.Abort(true, true, true, true, true); err != nil {
		return err
	}
	if err := stm.SWD.Select(0, 0); err != nil {
		return err
	}
	if err := stm.SWD.WriteCSW(CSW_ADDR_INCR_OFF, CSW_BITS_32); err != nil {
		return err
	}
	return err
}
