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

package logs

import (
    "fmt"
)

type Level int

const (
    DEBUG = iota
    INFO
    ERROR
)

type nameValue struct {
    Name string
    Value interface{}
}

// Args provides a way to safely pass additional params in a name=value format to the formatted log string.
type Args []nameValue

// Logger is the interface that wraps the basic log methods
type Logger interface {
    // Debug writes the string and args at DEBUG level
    Debug(str string, args ...Args)
    // Info writes the string and args at INFO level
    Info(str string, args ...Args)
    // Error writes the string and args at ERROR level
    Error(str string, args ...Args)
    // ErrorRet writes the string and args at ERROR level, adding error=err to the args.
    // It returns err, so its convenient to log the error and return it
    ErrorRet(err error, str string, args ...Args) error
    // Trace is used to write log message at upon entering and exiting the function.
    // Unlike other Logger methods, it does not get a string for message.
    // Instead, it writes only ENTER and EXIT
    // It writes the ENTER message when called, and returns a function that writes the EXIT message,
    // so it can be used with defer in 1 line
    Trace(level Level, args ...Args) func()
}

func (param nameValue) String() string {
    return fmt.Sprintf("{" + param.Name + "=%v}", param.Value)
}
