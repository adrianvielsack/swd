package swd

// Reference: https://developer.arm.com/documentation/dui0552/a/cortex-m3-peripherals/system-control-block?lang=en

const (
	//Application Interrupt and Reset Control Register
	AIRCR               = 0xE000ED0C
	AIRCR_VECTKEYSTAT   = 0x05FA0000
	AIRCR_ENDIANESS     = 1 << 15
	AIRCR_PRIGROUP_2    = 1 << 10
	AIRCR_PRIGROUP_1    = 1 << 9
	AIRCR_PRIGROUP_0    = 1 << 8
	AIRCR_SYSRESETREQ   = 1 << 2
	AIRCR_VECTCLRACTIVE = 1 << 1
	AIRCR_VECTRESET     = 1 << 1
)

func (cd *CoreDebug) SystemResetRequest() error {
	return cd.SWD.WriteRegister(AIRCR, AIRCR_VECTKEYSTAT|AIRCR_SYSRESETREQ)
}
