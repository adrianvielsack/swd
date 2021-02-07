package swd

import (
	"bytes"
	"errors"
)

func FlashSTM32(ctrl *STM32, data []byte) error {
	if err := ctrl.Initialize(); err != nil {
		return err
	}

	if err := ctrl.CoreDebug.Halt(); err != nil {
		return errors.New("Error halting core! Error: " + err.Error())
	}
	if err := ctrl.Flash.Writable(); err != nil {
		return errors.New("Error turning flash writable! Error: " + err.Error())
	}
	if err := ctrl.Flash.EraseAll(); err != nil {
		return NewFlashEraseError(err)
	}

	if err := ctrl.Flash.WriteAddress(0, data); err != nil {
		return NewFlashWriteError(err)
	}
	if readData, err := ctrl.Flash.Read(0, uint32(len(data))); err != nil {
		return err
	} else {
		if bytes.Compare(readData, data) != 0 {
			return &FlashVerifyError{}
		}
	}

	if err := ctrl.CoreDebug.RunAfterReset(); err != nil {
		return err
	}
	if err := ctrl.CoreDebug.SystemResetRequest(); err != nil {
		return err
	}

	return nil
}
