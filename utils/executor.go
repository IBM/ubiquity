package utils

import (
	"log"
	"os"
	"os/exec"
	"bytes"
	"fmt"
)

//go:generate counterfeiter -o ../fakes/fake_executor.go . Executor
type Executor interface { // basic host dependent functions
	Execute(command string, args []string) ([]byte, error)
	Stat(string) (os.FileInfo, error)
	Mkdir(string, os.FileMode) error
	RemoveAll(string) error
	Hostname() (string, error)
}

type executor struct {
	logger *log.Logger
}

func NewExecutor(logger *log.Logger) Executor {
	return &executor{logger: logger}
}

func (e *executor) Execute(command string, args []string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		e.logger.Printf("Error executing command: %#v, err: %s", cmd.Args,
			fmt.Sprint(err) + ": " + stderr.String())
		return nil, err
	}
	return stdout.Bytes(), err
}
func (e *executor) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (e *executor) Mkdir(path string, mode os.FileMode) error {
	return os.Mkdir(path, mode)
}

func (e *executor) RemoveAll(path string) error {

	return os.RemoveAll(path)
}

func (e *executor) Hostname() (string, error) {
	return os.Hostname()
}
