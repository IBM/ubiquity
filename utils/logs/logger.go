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

type Param struct {
    Name string
    Value interface{}
}

type Args []Param

type Logger interface {
    Debug(str string, args ...Args)
    Info(str string, args ...Args)
    Error(str string, args ...Args)
    ErrorRet(err error, str string, args ...Args) error
    Trace(level Level, args ...Args) func()
}

func (param Param) String() string {
    return fmt.Sprintf("{" + param.Name + "=%v}", param.Value)
}
