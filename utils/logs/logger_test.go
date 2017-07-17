package logs_test


import (
    "errors"
    "fmt"
    "github.com/IBM/ubiquity/utils/logs"
)

// Safely pass name=value pairs to be logged with the string.
func ExampleArgs() {
    logger := logs.GetLogger()
    logger.Info("the info message", logs.Args{{"name1", "value1"}, {"name2", "value2"}, {"name3", "value3"}})
    // (date) (time) INFO (PID) (filename)(line-number) (package-name)(function-name) the info message [[{name1=value1}] [{name2=value2}] [{name3=value3}]]
}

func ExampleLogger_errorRet() error {
    logger := logs.GetLogger()
    if err := errors.New("some-error"); err != nil {
        return logger.ErrorRet(err, "failed")
    }
    return nil
    // (date) (time) ERROR (PID) (filename)(line-number) (package-name)(function-name) failed [[{err=some-error}]]
}


func ExampleLogger_trace() {
    defer logs.GetLogger().Trace(logs.INFO)()
    fmt.Println("doing stuff")
    // (date) (time) INFO (PID) (filename)(line-number) (package-name)(function-name) ENTER []
    // doing stuff
    // (date) (time) INFO (PID) (filename)(line-number) (package-name)(function-name) EXIT []
}


