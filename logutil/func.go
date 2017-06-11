package logutil

import (
    "github.com/op/go-logging"
)


func (logger *impLogger) Debug(str string, param ...Param) {
    if logger.goLogger.IsEnabledFor(logging.DEBUG) {
        logger.goLogger.Debugf(str + " %v", param)
    }
}

func (logger *impLogger) Info(str string, param ...Param) {
    if logger.goLogger.IsEnabledFor(logging.INFO) {
        logger.goLogger.Infof(str + " %v", param)
    }
}

func (logger *impLogger) Error(str string, param ...Param) {
    if logger.goLogger.IsEnabledFor(logging.ERROR) {
        logger.goLogger.Errorf(str + " %v", param)
    }
}