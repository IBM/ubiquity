package logs_test

import (
    "github.com/IBM/ubiquity/utils/logs"
    "os"
    "fmt"
    "path"
)

var filePath = "/tmp/ubiquity.log"
var filePath2 = "/tmp/test/ubiquity.log"
var createFilePathSuccess = "Create file path success"
var createFilePathFail = "Create file path fail"

// Init a logger to file "/tmp/ubiquity.log" at DEBUG level
func ExampleInitFileLogger() {
	//Example Test, shows how to use the logs.InitFileLogger
    defer logs.InitFileLogger(logs.DEBUG, filePath, 50)()
    if _, err := os.Stat(filePath); err == nil{
        fmt.Println(createFilePathSuccess)
    } else {
        fmt.Println(createFilePathFail)
    }
    // Output: Create file path success
}

// Init a logger to file "/tmp/test/ubiquity.log" at DEBUG level, dir exist before init
func ExampleInitFileLogger2() {
	//Example Test, shows how to use the logs.InitFileLogger
    _, err := os.Stat(filePath2)
	if os.IsNotExist(err) {
		fileDir,_ := path.Split(filePath2)
		err := os.MkdirAll(fileDir, 0766)
		if err != nil {
			panic(fmt.Sprintf("failed to create log folder %v", err))
		}
	}
    defer logs.InitFileLogger(logs.DEBUG, filePath2, 50)()
    if _, err := os.Stat(filePath2); err == nil{
        fmt.Println(createFilePathSuccess)
    } else {
        fmt.Println(createFilePathFail)
    }
    // Output: Create file path success
}

// Init a logger to stdout at DEBUG level
func ExampleInitStdoutLogger() {
    defer logs.InitStdoutLogger(logs.DEBUG)()
}