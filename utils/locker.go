package utils

import (
	"log"
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

func NewLocker(logger *log.Logger) Locker {
	return &locker{locks: make(map[string]*sync.RWMutex), accessLock: &sync.Mutex{}, statsLock: &sync.Mutex{}, cleanupLock: &sync.Mutex{}, stats: make(map[string]time.Time), logger: logger}
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
	logger      *log.Logger
}

func (l *locker) WriteLock(name string) {
	l.logger.Printf("WriteLock start for '%s'\n", name)
	defer l.logger.Printf("WriteLock end for '%s'\n", name)

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
	l.logger.Printf("WriteUnlock start for '%s'\n", name)
	defer l.logger.Printf("WriteUnlock end for '%s'\n", name)
	defer l.updateStats(name)
	l.accessLock.Lock()
	defer l.accessLock.Unlock()
	if lock, exists := l.locks[name]; exists {
		lock.Unlock()
		return
	}

}
func (l *locker) ReadLock(name string) {
	l.logger.Printf("ReadLock start for %s\n", name)
	defer l.logger.Printf("ReadLock end for '%s'\n", name)
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
	l.logger.Printf("ReadUnlock start for '%s'\n", name)
	defer l.logger.Printf("ReadUnlock end for '%s'\n", name)
	defer l.updateStats(name)
	l.accessLock.Lock()
	defer l.accessLock.Unlock()
	if lock, exists := l.locks[name]; exists {
		lock.RUnlock()
		return
	}
}
func (l *locker) updateStats(name string) {
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
	l.cleanupLock.Lock()
	defer l.cleanupLock.Unlock()
	currentTime := time.Now()
	var statsToDelete []string
	for name, stat := range l.stats {
		if currentTime.Sub(stat).Seconds() > STALE_LOCK_TIMEOUT {
			l.logger.Printf("Removing stalelock '%s' as it has exceeded configured timeout ('%d seconds')\n", name, STALE_LOCK_TIMEOUT)
			delete(l.locks, name)
			statsToDelete = append(statsToDelete, name)
		}
	}
	for _, name := range statsToDelete {
		delete(l.stats, name)
	}
}
