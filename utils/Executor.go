package utils

import (
	"log"
	"os/exec"
)

//go:generate counterfeiter -o ../fakes/fake_executor.go . Executor
type Executor interface {
	Execute(command string, args []string) ([]byte, error)
}

type executor struct {
	logger *log.Logger
}

func NewExecutor(logger *log.Logger) Executor {
	return &executor{logger: logger}
}

func (e *executor) Execute(command string, args []string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		e.logger.Printf("Error executing command %v", err)
		return nil, err
	}
	return output, err
}
