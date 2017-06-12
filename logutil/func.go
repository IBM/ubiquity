package logutil

import (
    "github.com/op/go-logging"
    "fmt"
)


func (logger *impLogger) Debug(str string, args ...Args) {
    if logger.goLogger.IsEnabledFor(logging.DEBUG) {
        logger.goLogger.Debugf(str + " %v", args)
    }
}

func (logger *impLogger) Info(str string, args ...Args) {
    if logger.goLogger.IsEnabledFor(logging.INFO) {
        logger.goLogger.Infof(str + " %v", args)
    }
}

func (logger *impLogger) Error(str string, args ...Args) {
    if logger.goLogger.IsEnabledFor(logging.ERROR) {
        logger.goLogger.Errorf(str + " %v", args)
    }
}

func (param Param) String() string {
    return fmt.Sprintf("{" + param.Name + "=%v}", param.Value)
}
