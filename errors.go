package swd

type FlashVerifyError struct {
}

func (f FlashVerifyError) Error() string {
	return "Error verifying Flash memory!"
}

type FlashEraseError struct {
	err error
}

func (f FlashEraseError) Error() string {
	return "Error erasing flash! Error: " + f.err.Error()
}

func NewFlashEraseError(err error) *FlashEraseError {
	return &FlashEraseError{err: err}
}

type FlashWriteError struct {
	err error
}

func (f FlashWriteError) Error() string {
	return "Error writing flash! Error: " + f.err.Error()
}

func NewFlashWriteError(err error) *FlashWriteError {
	return &FlashWriteError{err: err}
}
