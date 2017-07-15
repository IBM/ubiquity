package logs

import (
    "errors"
    "fmt"
)

// Safely pass name=value pairs to be logged with the string.
func ExampleArgs() {
    logger = GetLogger()
    logger.Info("the info message", Args{{"name1", "value1"}, {"name2", "value2"}, {"name3", "value3"}})
    // Output: (date) (time) INFO (PID) (filename)(line-number) (package-name)(function-name) the info message [[{name1=value1}] [{name2=value2}] [{name3=value3}]]
}

func ExampleLogger_errorRet() error {
    logger = GetLogger()
    err := errors.New("some-error")
    if err != nil {
        return logger.ErrorRet(err, "failed")
    }
    // Output: (date) (time) ERROR (PID) (filename)(line-number) (package-name)(function-name) failed [[{err=some-error}]]
}


func ExampleLogger_trace() {
    defer GetLogger().Trace(INFO)()
    fmt.Println("doing stuff")
    // Output: (date) (time) INFO (PID) (filename)(line-number) (package-name)(function-name) ENTER []
    // doing stuff
    // (date) (time) INFO (PID) (filename)(line-number) (package-name)(function-name) EXIT []
}