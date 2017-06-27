package utils

import (
	"bytes"
	"github.com/IBM/ubiquity/utils/logs"
	"os"
	"os/exec"
)

//go:generate counterfeiter -o ../fakes/fake_executor.go . Executor
type Executor interface { // basic host dependent functions
	Execute(command string, args []string) ([]byte, error)
	Stat(string) (os.FileInfo, error)
	Mkdir(string, os.FileMode) error
	MkdirAll(string, os.FileMode) error
	RemoveAll(string) error
	Hostname() (string, error)
	IsExecutable(string) error
}

type executor struct {
	logger logs.Logger
}

func NewExecutor() Executor {
	return &executor{logs.GetLogger()}
}

func (e *executor) Execute(command string, args []string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	stdErr := stderr.Bytes()
	stdOut := stdout.Bytes()
	e.logger.Debug(
		"Command executed with args and error and output.",
		logs.Args{
			{"command", command},
			{"args", args},
			{"error", string(stdErr[:])},
			{"output", string(stdOut[:])},
		})

	if err != nil {
		return nil, err
	}
	return stdOut, err
}
func (e *executor) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (e *executor) Mkdir(path string, mode os.FileMode) error {
	return os.Mkdir(path, mode)
}

func (e *executor) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func (e *executor) RemoveAll(path string) error {

	return os.RemoveAll(path)
}

func (e *executor) Hostname() (string, error) {
	return os.Hostname()
}

func (e *executor) IsExecutable(path string) error {
	_, err := exec.LookPath(path)
	return err
}
