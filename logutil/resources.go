package logutil

import (
    "github.com/op/go-logging"
)

const (
    DEBUG = logging.DEBUG
    INFO = logging.INFO
    ERROR = logging.ERROR
)

type Level logging.Level

type Param struct {
    Name string
    Value interface{}
}

type Logger interface {
    Debug(str string, param ...Param)
    Info(str string, param ...Param)
    Error(str string, param ...Param)
}

type impLogger struct {
    goLogger    *logging.Logger
}