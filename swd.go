package swd

const (
	SWD_DP_IDCODE    byte = 0x00
	SWD_DP_ABORT     byte = 0x00
	SWD_DP_CTRL_STAT byte = 0x04
	SWD_DP_RESEND    byte = 0x08
	SWD_DP_SELECT    byte = 0x08
	SWD_DP_RDBUFF    byte = 0x0C

	CTL_CSYSPWRUPREQ = 0x40000000
	CTL_CDBGPWRUPREQ = 0x10000000
	CTL_CDBGRSTREQ   = 0x04000000

	MEMAP_CSW = 0x00
	MEMAP_TAR = 0x04
	MEMAP_DRW = 0x0C

	CSW_BITS_8  = 0b000
	CSW_BITS_16 = 0b001
	CSW_BITS_32 = 0b010

	CSW_ADDR_INCR_OFF    = 0b00
	CSW_ADDR_INCR_SINGLE = 0b01
	CSW_ADDR_INCR_PACKET = 0b10

	CSW_MASTER_DEBUG = 1 << 29
	CSW_HPROT1       = 1 << 25
)

type SWD struct {
	swd        SWDLowLevel
	accessPort uint8
	bank       uint8
}

func NewSWD(swd SWDLowLevel) *SWD {
	return &SWD{
		swd:        swd,
		bank:       0,
		accessPort: 0,
	}
}

func (s *SWD) PowerOnRequest() error {
	return s.swd.SWDWrite(false, SWD_DP_CTRL_STAT, CTL_CDBGPWRUPREQ|CTL_CDBGRSTREQ|CTL_CSYSPWRUPREQ)
}

func (s *SWD) Select(accessPort uint8, bank uint8) error {
	s.bank = bank & 0xF0
	return s.swd.SWDWrite(false, SWD_DP_SELECT, (uint32(accessPort)<<24)|(uint32(s.bank)>>4))
}

func (s *SWD) writeMemAP(address uint8, data uint32) error {
	if s.bank != (address & 0xF0) {
		if err := s.Select(0, address); err != nil {
			return err
		}
	}

	return s.swd.SWDWrite(true, address&0xF, data)
}

func (s *SWD) ReadRDBuff() (uint32, error) {
	return s.swd.SWDRead(false, SWD_DP_RDBUFF)
}

func (s *SWD) readMemAP(address uint8) (uint32, error) {
	if s.bank != (address & 0xF0) {
		if err := s.Select(0, address); err != nil {
			return 0, err
		}
	}

	_, err := s.swd.SWDRead(true, address)
	if err != nil {
		return 0, err
	}
	return s.ReadRDBuff()
}

func (s *SWD) WriteCSW(addrIncrease uint8, writeSize uint8) error {
	val, err := s.readMemAP(MEMAP_CSW)
	if err != nil {
		return err
	}

	return s.writeMemAP(MEMAP_CSW, val&0xFFFFFF00|(((uint32(addrIncrease)&0x3)<<4)|(uint32(writeSize)&0x07)))
}

func (s *SWD) WriteTAR(address uint32) error {
	return s.writeMemAP(MEMAP_TAR, address)
}

func (s *SWD) WriteDRW(data uint32) error {
	return s.writeMemAP(MEMAP_DRW, data)
}

func (s *SWD) ReadDRW() (uint32, error) {
	return s.readMemAP(MEMAP_DRW)
}

func (s *SWD) Abort(overrunnerr bool, wdataerr bool, stickyerr bool, stickycmp bool, dapabort bool) error {
	reg := uint32(0)
	if overrunnerr {
		reg |= 1 << 4
	}
	if wdataerr {
		reg |= 1 << 3
	}
	if stickyerr {
		reg |= 1 << 2
	}
	if stickycmp {
		reg |= 1 << 1
	}
	if dapabort {
		reg |= 1 << 0
	}

	return s.swd.SWDWrite(false, SWD_DP_ABORT, reg)
}

func (s *SWD) WriteRegister(register uint32, data uint32) error {
	err := s.WriteTAR(register)
	if err != nil {
		return err
	}

	return s.WriteDRW(data)
}

func (s *SWD) WriteRegisterHalf(register uint32, data uint32) error {
	err := s.WriteTAR(register)
	if err != nil {
		return err
	}
	if err := s.WriteCSW(CSW_ADDR_INCR_OFF, CSW_BITS_16); err != nil {
		return err
	}
	if err := s.WriteDRW(data); err != nil {
		return err
	}
	return s.WriteCSW(CSW_ADDR_INCR_OFF, CSW_BITS_32)
}

func (s *SWD) ReadRegister(register uint32) (uint32, error) {
	err := s.WriteTAR(register)
	if err != nil {
		return 0, err
	}

	return s.ReadDRW()
}

func (s *SWD) ReadID() (uint32, error) {
	return s.swd.SWDRead(false, SWD_DP_IDCODE)
}

func (s *SWD) ReadAppID() (uint32, error) {
	return s.readMemAP(0xFC)
}
