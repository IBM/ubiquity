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
	"bytes"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"sync"

	"github.com/IBM/ubiquity/resources"
	"github.com/op/go-logging"
	"k8s.io/apimachinery/pkg/util/uuid"

)

const (
	traceEnter = "ENTER"
	traceExit  = "EXIT"
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

func GetGoID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

var GoIdToRequestIdMap = new(sync.Map)

func (l *goLoggingLogger) getContextStringFromGoid() string {
	go_id := GetGoID()
	context, exists := GoIdToRequestIdMap.Load(go_id)
	if !exists {
		context = resources.RequestContext{Id: "NA"}
	} else {
		context = context.(resources.RequestContext)
	}
	return fmt.Sprintf("%s:%d", context.(resources.RequestContext).Id, go_id)

}

func GetDeleteFromMapFunc(key interface{}) func() {
	return func() { GoIdToRequestIdMap.Delete(key) }
}

func (l *goLoggingLogger) Debug(str string, args ...Args) {
	goid_context_string := l.getContextStringFromGoid()
	l.logger.Debugf(fmt.Sprintf("[%s] %s %v", goid_context_string, str, args))
}

func (l *goLoggingLogger) Info(str string, args ...Args) {
	goid_context_string := l.getContextStringFromGoid()
	l.logger.Infof(fmt.Sprintf("[%s] %s %v", goid_context_string, str, args))
}

func (l *goLoggingLogger) Error(str string, args ...Args) {
	goid_context_string := l.getContextStringFromGoid()
	l.logger.Errorf(fmt.Sprintf("[%s] %s %v", goid_context_string, str, args))
}

func (l *goLoggingLogger) ErrorRet(err error, str string, args ...Args) error {
	goid_context_string := l.getContextStringFromGoid()
	l.logger.Errorf(fmt.Sprintf("[%s] %s %v", goid_context_string, str, append(args, Args{{"error", err}})))
	return err
}

func (l *goLoggingLogger) Trace(level Level, args ...Args) func() {
	goid_context_string := l.getContextStringFromGoid()
	log_string_enter := fmt.Sprintf("[%s] %s", goid_context_string, traceEnter)
	log_string_exit := fmt.Sprintf("[%s] %s", goid_context_string, traceExit)
	switch level {
	case DEBUG:
		l.logger.Debug(log_string_enter, args)
		return func() { l.logger.Debug(log_string_exit, args) }
	case INFO:
		l.logger.Info(log_string_enter, args)
		return func() { l.logger.Info(log_string_exit, args) }
	case ERROR:
		l.logger.Error(log_string_enter, args)
		return func() { l.logger.Error(log_string_exit, args) }
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

func GetNewRequestContext() resources.RequestContext{
	request_uuid := fmt.Sprintf("%s", uuid.NewUUID())
    return resources.RequestContext{Id: request_uuid}
}

