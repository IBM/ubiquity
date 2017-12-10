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

package utils

import (
	"bytes"
	"github.com/IBM/ubiquity/utils/logs"
	"os"
	"os/exec"
	"path/filepath"
	"context"
	"time"
	"fmt"
)

//go:generate counterfeiter -o ../fakes/fake_executor.go . Executor
type Executor interface { // basic host dependent functions
	Execute(command string, args []string) ([]byte, error)
	Stat(string) (os.FileInfo, error)
	Mkdir(string, os.FileMode) error
	MkdirAll(string, os.FileMode) error
	RemoveAll(string) error
	Remove(string) error
	Hostname() (string, error)
	IsExecutable(string) error
	IsNotExist(error) bool
	EvalSymlinks(path string) (string, error)
	ExecuteWithTimeout(mSeconds int ,command string, args []string) ([]byte, error)
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

	return stdOut, err
}

func (e *executor) ExecuteWithTimeout(mSeconds int ,command string, args []string) ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Start()
	e.logger.Debug(fmt.Sprintf("Command %s Started", command))
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()
	for i := 0; i < 5; i++ {
		select {
		case err := <-done:
			e.logger.Debug(fmt.Sprintf("Command %s Done: %s", command, err))
			stdErr := stderr.Bytes()
			stdOut := stdout.Bytes()
			e.logger.Debug(
				"Command executed with args and error and output.",
				logs.Args{
					{"command", command},
					{"args", args},
					{"error", string(stdErr[:])},
					{"output", string(stdOut[:])},
					{"timeout mSeconds", mSeconds},
					{"exit status", err.Error()},
				})

			return stdOut, err
		case <-time.After(time.Duration(mSeconds) * time.Millisecond):
			e.logger.Debug(fmt.Sprintf("Command %s reched timeout after: %d", command, mSeconds))
			stdErr := stderr.Bytes()
			stdOut := stdout.Bytes()
			e.logger.Debug(
				"Command executed with args and error and output.",
				logs.Args{
					{"command", command},
					{"args", args},
					{"error", string(stdErr[:])},
					{"output", string(stdOut[:])},
					{"timeout mSeconds", mSeconds},
					{"exit status", err.Error()},
				})
			return stdOut, err
		default:
			e.logger.Debug(fmt.Sprintf("waiting for cmd %s", command))
			time.Sleep(1000*time.Millisecond)
		}
	}
	return stdout.Bytes(), err
}

func (e *executor) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (e *executor) IsNotExist(err error) bool{
	return os.IsNotExist(err)
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
func (e *executor) Remove(path string) error {
	return os.Remove(path)
}

func (e *executor) Hostname() (string, error) {
	return os.Hostname()
}

func (e *executor) IsExecutable(path string) error {
	_, err := exec.LookPath(path)
	return err
}

func (e *executor) EvalSymlinks(path string) (string, error) {
	evalSlink, err := filepath.EvalSymlinks(path)
	return evalSlink, err
}


