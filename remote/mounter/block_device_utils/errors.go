package block_device_utils

import "fmt"

type commandNotFoundError struct {
	cmd string
	err error
}

func (e *commandNotFoundError) Error() string {
	return fmt.Sprintf("command [%s] not found [%s]", e.cmd, e.err)
}

type commandExecuteError struct {
	cmd string
	err error
}

func (e *commandExecuteError) Error() string {
	return fmt.Sprintf("command [%s] execute failed [%s]", e.cmd, e.err)
}

type volumeNotFoundError struct {
	volName string
}

func (e *volumeNotFoundError) Error() string {
	return fmt.Sprintf("volume [%s] not found", e.volName)
}

type unsupportedProtocolError struct {
	protocol Protocol
}

func (e *unsupportedProtocolError) Error() string {
	return fmt.Sprintf("unsupported protocol [%v]", e.protocol)
}
