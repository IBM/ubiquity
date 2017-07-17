package logs_test

import "github.com/IBM/ubiquity/utils/logs"

// Init a logger to file "/tmp/ubiquity.log" at DEBUG level
func ExampleInitFileLogger() {
    defer logs.InitFileLogger(logs.DEBUG, "/tmp/ubiquity.log")()
}

// Init a logger to stdout at DEBUG level
func ExampleInitStdoutLogger() {
    defer logs.InitStdoutLogger(logs.DEBUG)()
}
