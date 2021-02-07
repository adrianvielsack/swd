package swd

type SWDLowLevel interface {
	Reset()
	Sync()
	SWDWrite(ap bool, cmd byte, data uint32) error
	SWDRead(ap bool, address byte) (uint32, error)
	Initialize()
}
