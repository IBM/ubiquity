package logutil

import (
    "fmt"
    "github.com/op/go-logging"
)

const (
    traceEnter = "ENTER"
    traceExit = "EXIT"
)

func (logger *impLogger) Debug(str string, args ...Args) {
    logger.goLogger.Debugf(str + " %v", args)
}

func (logger *impLogger) Info(str string, args ...Args) {
    logger.goLogger.Infof(str + " %v", args)
}

func (logger *impLogger) Error(str string, args ...Args) {
    logger.goLogger.Errorf(str + " %v", args)
}

func (logger *impLogger) ErrorRet(err error, str string, args ...Args) error {
    logger.goLogger.Errorf(str + " %v ", append(args, Args{{"error", err}}))
    return err
}


func (logger *impLogger) Trace(level logging.Level, args ...Args) func() {
    switch level {
    case DEBUG:
        logger.goLogger.Debug(traceEnter, args)
        return func() { logger.goLogger.Debug(traceExit, args) }
    case INFO:
        logger.goLogger.Info(traceEnter, args)
        return func() { logger.goLogger.Info(traceExit, args) }
    case ERROR:
        logger.goLogger.Error(traceEnter, args)
        return func() { logger.goLogger.Error(traceExit, args) }
    default:
        panic("unknown level")
    }
}


func (param Param) String() string {
    return fmt.Sprintf("{" + param.Name + "=%v}", param.Value)
}
