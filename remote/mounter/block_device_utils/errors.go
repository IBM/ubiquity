/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package block_device_utils

import "fmt"

type commandNotFoundError struct {
	cmd string
	err error
}

func (e *commandNotFoundError) Error() string {
	return fmt.Sprintf("command [%v] is not found [%v]", e.cmd, e.err)
}

type commandExecuteError struct {
	cmd string
	err error
}

func (e *commandExecuteError) Error() string {
	return fmt.Sprintf("command [%v] execution failure [%v]", e.cmd, e.err)
}

type volumeNotFoundError struct {
	volName string
}

func (e *volumeNotFoundError) Error() string {
	return fmt.Sprintf("volume [%v] is not found", e.volName)
}

type unsupportedProtocolError struct {
	protocol Protocol
}

func (e *unsupportedProtocolError) Error() string {
	return fmt.Sprintf("Protocol [%v] is not supported", e.protocol)
}
