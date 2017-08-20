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
    "github.com/op/go-logging"
    "io"
)

const (
    traceEnter = "ENTER"
    traceExit = "EXIT"
)

type goLoggingLogger struct {
    logger *logging.Logger
}

func newGoLoggingLogger(level Level, writer io.Writer) *goLoggingLogger {
    newLogger := logging.MustGetLogger("")
    newLogger.ExtraCalldepth = 1
    format := logging.MustStringFormatter("%{time:2006-01-02 15:04:05.999} %{level:.5s} %{pid} %{shortfile} %{shortpkg}::%{shortfunc} %{message}")
    backend := logging.NewLogBackend(writer, "", 0)
    backendFormatter := logging.NewBackendFormatter(backend, format)
    backendLeveled := logging.AddModuleLevel(backendFormatter)
    backendLeveled.SetLevel(getLevel(level), "")
    newLogger.SetBackend(backendLeveled)
    return &goLoggingLogger{newLogger}
}

func (l *goLoggingLogger) Debug(str string, args ...Args) {
    l.logger.Debugf(str + " %v", args)
}

func (l *goLoggingLogger) Info(str string, args ...Args) {
    l.logger.Infof(str + " %v", args)
}

func (l *goLoggingLogger) Error(str string, args ...Args) {
    l.logger.Errorf(str + " %v", args)
}

func (l *goLoggingLogger) ErrorRet(err error, str string, args ...Args) error {
    l.logger.Errorf(str + " %v ", append(args, Args{{"error", err}}))
    return err
}

func (l *goLoggingLogger) Trace(level Level, args ...Args) func() {
    switch level {
    case DEBUG:
        l.logger.Debug(traceEnter, args)
        return func() { l.logger.Debug(traceExit, args) }
    case INFO:
        l.logger.Info(traceEnter, args)
        return func() { l.logger.Info(traceExit, args) }
    case ERROR:
        l.logger.Error(traceEnter, args)
        return func() { l.logger.Error(traceExit, args) }
    default:
        panic("unknown level")
    }
}

func getLevel(level Level) logging.Level {
    switch level {
    case DEBUG:
        return logging.DEBUG
    case INFO:
        return logging.INFO
    case ERROR:
        return logging.ERROR
    default:
        panic("unknown level")
    }
}
