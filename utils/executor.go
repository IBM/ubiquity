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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/IBM/ubiquity/utils/logs"
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
	ExecuteWithTimeout(mSeconds int, command string, args []string) ([]byte, error)
	Lstat(path string) (os.FileInfo, error)
	IsDir(fInfo os.FileInfo) bool
	Symlink(target string, slink string) error
	IsSlink(fInfo os.FileInfo) bool
	GetGlobFiles(file_pattern string) (matches []string, err error)
	IsSameFile(file1 os.FileInfo, file2 os.FileInfo) bool
	IsDirEmpty(dir string) (bool, error)
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

func (e *executor) ExecuteWithTimeout(mSeconds int, command string, args []string) ([]byte, error) {

	// Create a new context and add a timeout to it
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mSeconds)*time.Millisecond)
	defer cancel() // The cancel should be deferred so resources are cleaned up

	// Create the command with our context
	cmd := exec.CommandContext(ctx, command, args...)

	// This time we can simply use Output() to get the result.
	out, err := cmd.Output()

	// We want to check the context error to see if the timeout was executed.
	// The error returned by cmd.Output() will be OS specific based on what
	// happens when a process is killed.
	if ctx.Err() == context.DeadlineExceeded {
		e.logger.Debug(fmt.Sprintf("Command %s timeout reached", command))
		return nil, ctx.Err()
	}

	e.logger.Debug(fmt.Sprintf("Yixuan Command is %s", command))
	e.logger.Debug(fmt.Sprintf("Yixuan whole Command is %v", cmd))
	// If there's no context error, we know the command completed (or errored).
	e.logger.Debug(fmt.Sprintf("Output from command: %s", string(out)))
	if err != nil {
		e.logger.Debug(fmt.Sprintf("Non-zero exit code: %s", err))
	}

	return out, err
}

func (e *executor) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (e *executor) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

func (e *executor) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (e *executor) IsDir(fInfo os.FileInfo) bool {
	return fInfo.IsDir()
}

func (e *executor) IsSlink(fInfo os.FileInfo) bool {
	return fInfo.Mode()&os.ModeSymlink != 0
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

func (e *executor) Symlink(target string, slink string) error {
	return os.Symlink(target, slink)
}

func (e *executor) GetGlobFiles(file_pattern string) (matches []string, err error) {
	return filepath.Glob(file_pattern)
}
func (e *executor) IsSameFile(file1 os.FileInfo, file2 os.FileInfo) bool {
	return os.SameFile(file1, file2)
}

func (e *executor) IsDirEmpty(dir string) (bool, error) {
	files, err := ioutil.ReadDir(dir)
	e.logger.Debug("the files", logs.Args{{"files", files}})
	if err != nil {
		return false, err
	}

	return len(files) == 0, nil
}
