package swd

import "log"

const (
	//For Cortex M3 e.g. STM32F1
	//Ref: https://developer.arm.com/documentation/ddi0337/e/core-debug/core-debug-registers

	//Debug Halting Control and Status Register
	DHCSR             = 0xE000EDF0
	DHCSR_DBGKEY      = 0xA05F0000
	DHCSR_S_RESET_ST  = 1 << 25
	DHCSR_S_RETIRE_ST = 1 << 24
	DHCSR_S_LOCKUP    = 1 << 19
	DHCSR_S_SLEEP     = 1 << 18
	DHCSR_S_HALT      = 1 << 17
	DHCSR_S_REGRDY    = 1 << 16
	DHCSR_C_SNAP_ALL  = 1 << 5
	DHCSR_C_MASKINTS  = 1 << 3
	DHCSR_C_STEP      = 1 << 2
	DHCSR_C_HALT      = 1 << 1
	DHCSR_C_DEBUGEN   = 1 << 0

	//Debug Core Register Selector Register
	DCRSR        = 0xE000EDF4
	DCRSR_REGWnR = 1 << 16

	//Debug Core Register Data Register
	DCRDR = 0xE000EDF8

	//Debug Exception and Monitor Control Register
	DEMCR                     = 0xE000EDFC
	DEMCR_TRCENA              = 1 << 24
	DEMCR_MON_REQ             = 1 << 19
	DEMCR_MON_STEP            = 1 << 18
	DEMCR_MON_PEND            = 1 << 17
	DEMCR_MON_EN              = 1 << 16
	DEMCR_VC_HARDERR          = 1 << 10
	DEMCR_VC_INTERR           = 1 << 9
	DEMCR_VC_BUSERR           = 1 << 8
	DEMCR_VC_STATERR          = 1 << 7
	DEMCR_VC_VC_CHKERR        = 1 << 6
	DEMCR_VC_NOCPERRDEMCR_VC_ = 1 << 5
	DEMCR_VC_MMERR            = 1 << 4
	DEMCR_VC_CORERESET        = 1 << 0
)

type CoreDebug struct {
	*SWD
}

func (cd *CoreDebug) ResetRegisters() error {
	if err := cd.WriteRegister(DHCSR, DHCSR_DBGKEY); err != nil {
		return err
	}
	if err := cd.WriteRegister(DEMCR, 0); err != nil {
		return err
	}
	return cd.SystemResetRequest()
}

func (cd *CoreDebug) Halt() error {
	if err := cd.WriteRegister(DHCSR, DHCSR_DBGKEY|DHCSR_C_DEBUGEN|DHCSR_C_HALT); err != nil {
		return err
	}
	var err error
	var reg uint32

	for n := 0; n < 3; n++ {
		reg, err = cd.ReadRegister(DHCSR)
		log.Printf("Reading DHCSR = %08X\n", reg)
		if err != nil {
			log.Println("Error reading Status register DHCSR")
		}

		if reg&DHCSR_S_HALT == DHCSR_S_HALT {
			return nil
		}
	}
	return err
}

func (cd *CoreDebug) Continue() error {
	if err := cd.WriteRegister(DHCSR, DHCSR_DBGKEY|DHCSR_C_DEBUGEN); err != nil {
		return err
	}
	var err error
	var reg uint32
	for n := 0; n < 3; n++ {
		reg, err = cd.ReadRegister(DHCSR)

		if err != nil {
			log.Println("Error reading Status register DHCSR")
			log.Printf("Reading DHCSR = %08X\n", reg)
		}

		if reg&DHCSR_S_HALT == 0 {
			return nil
		}
	}
	return err
}

func (cd *CoreDebug) RunAfterReset() error {
	if err := cd.WriteRegister(DHCSR, DHCSR_DBGKEY); err != nil {
		return err
	}
	var err error
	var reg uint32
	for n := 0; n < 3; n++ {
		reg, err = cd.ReadRegister(DHCSR)
		if err != nil {
			log.Println("Error reading Status register DHCSR")
			log.Printf("Reading DHCSR = %08X\n", reg)
		}

		if reg&DHCSR_S_HALT == 0 {
			return nil
		}
	}
	return err
}
