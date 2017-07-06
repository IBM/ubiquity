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
	"fmt"
	"github.com/IBM/ubiquity/utils/logs"
	"sync"
	"time"
)

//go:generate counterfeiter -o ../fakes/fake_locker.go . Locker
type Locker interface {
	WriteLock(name string)
	WriteUnlock(name string)
	ReadLock(name string)
	ReadUnlock(name string)
}

func NewLocker() Locker {
	return &locker{locks: make(map[string]*sync.RWMutex), accessLock: &sync.Mutex{}, statsLock: &sync.Mutex{}, cleanupLock: &sync.Mutex{}, stats: make(map[string]time.Time), logger: logs.GetLogger()}
}

const (
	STALE_LOCK_TIMEOUT = 600 //in seconds
)

type locker struct {
	accessLock  *sync.Mutex
	locks       map[string]*sync.RWMutex
	statsLock   *sync.Mutex
	stats       map[string]time.Time
	cleanupLock *sync.Mutex
	logger      logs.Logger
}

func (l *locker) WriteLock(name string) {
	defer l.logger.Trace(logs.DEBUG, logs.Args{{"lockName", name}})()

	defer l.updateStats(name)
	l.accessLock.Lock()
	if lock, exists := l.locks[name]; exists {
		l.accessLock.Unlock()
		lock.Lock()
		return
	}

	lock := &sync.RWMutex{}
	lock.Lock()
	l.locks[name] = lock
	l.accessLock.Unlock()
}
func (l *locker) WriteUnlock(name string) {
	defer l.logger.Trace(logs.DEBUG, logs.Args{{"lockName", name}})()
	defer l.updateStats(name)
	l.accessLock.Lock()
	defer l.accessLock.Unlock()
	if lock, exists := l.locks[name]; exists {
		lock.Unlock()
		return
	}

}
func (l *locker) ReadLock(name string) {
	defer l.logger.Trace(logs.DEBUG, logs.Args{{"lockName", name}})()
	defer l.updateStats(name)
	l.accessLock.Lock()
	if lock, exists := l.locks[name]; exists {
		l.accessLock.Unlock()
		lock.RLock()
		return
	}

	lock := &sync.RWMutex{}
	lock.RLock()
	l.locks[name] = lock
	l.accessLock.Unlock()
}
func (l *locker) ReadUnlock(name string) {
	defer l.logger.Trace(logs.DEBUG, logs.Args{{"lockName", name}})()
	defer l.updateStats(name)
	l.accessLock.Lock()
	defer l.accessLock.Unlock()
	if lock, exists := l.locks[name]; exists {
		lock.RUnlock()
		return
	}
}
func (l *locker) updateStats(name string) {
	defer l.logger.Trace(logs.DEBUG, logs.Args{{"lockName", name}})()

	l.statsLock.Lock()
	defer l.cleanup()
	defer l.statsLock.Unlock()
	if stat, exists := l.stats[name]; exists {
		stat = time.Now()
		l.stats[name] = stat
		return
	}
	stat := time.Now()
	l.stats[name] = stat
}
func (l *locker) cleanup() {
	defer l.logger.Trace(logs.DEBUG)()

	l.cleanupLock.Lock()
	defer l.cleanupLock.Unlock()
	currentTime := time.Now()
	var statsToDelete []string
	for name, stat := range l.stats {
		if currentTime.Sub(stat).Seconds() > STALE_LOCK_TIMEOUT {
			msg := fmt.Sprint("Removing stalelock '%s' as it has exceeded configured timeout ('%d seconds')\n", name, STALE_LOCK_TIMEOUT)
			l.logger.Debug(msg)
			delete(l.locks, name)
			statsToDelete = append(statsToDelete, name)
		}
	}
	for _, name := range statsToDelete {
		delete(l.stats, name)
	}
}
