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
