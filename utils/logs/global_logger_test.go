package logs

// Init a logger to file "/tmp/ubiquity.log" at DEBUG level
func ExampleInitFileLogger() {
    defer InitFileLogger(DEBUG, "/tmp/ubiquity.log")()
}

// Init a logger to stdout at DEBUG level
func ExampleInitStdoutLogger() {
    defer InitStdoutLogger(DEBUG)()
}
