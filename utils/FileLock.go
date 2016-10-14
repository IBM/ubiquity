package utils

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/djherbis/times"
)

type FileLock struct {
	Filesystem string
	Mountpoint string
	log        *log.Logger
}

const (
	LOCK_STALE_TIMEOUT  time.Duration = time.Second * 60
	LOCK_RETRY_TIMEOUT  time.Duration = time.Second * 60
	LOCK_RETRY_INTERVAL time.Duration = time.Second * 5
)

func NewFileLock(log *log.Logger, filesystem, mountpoint string) *FileLock {
	return &FileLock{log: log, Filesystem: filesystem, Mountpoint: mountpoint}
}

func (l *FileLock) Lock() error {

	var sleepTime time.Duration
	var attempt int

	lockFile := "spectrum-scale-" + l.Filesystem + ".lock"
	lockPath := path.Join(l.Mountpoint, lockFile)

	for sleepTime < LOCK_RETRY_TIMEOUT {

		attempt++
		str := fmt.Sprintf("Attempt %v to acquire lock using lockPath %s", attempt, lockPath)
		l.log.Println(str)

		fd, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL, 0700)

		if err != nil {
			if os.IsExist(err) {
				l.log.Printf("%v already exists", lockPath)

				fi, err := times.Stat(lockPath)

				if err != nil {
					if os.IsNotExist(err) {
						continue
					} else {
						return fmt.Errorf("Failed to stat %s : %s\n", lockPath, err.Error())
					}
				}
				ctime := fi.ChangeTime()

				if time.Since(ctime) >= LOCK_STALE_TIMEOUT {
					l.log.Printf("Found stale lock file : %s\n", lockPath)

					err := os.Remove(lockPath)

					if err != nil {
						if os.IsNotExist(err) {
							continue
						} else {
							return fmt.Errorf("Failed to delete stale lock file %s\n", lockPath)
						}
					} else {
						l.log.Printf("Successfully deleted stale lock file %s\n", lockPath)
					}
					continue
				}

				start := time.Now()
				time.Sleep(LOCK_RETRY_INTERVAL)
				sleepTime += time.Since(start)

			} else {
				return fmt.Errorf("Failed to create lock file %s : %s\n", lockPath, err.Error())
			}
		}

		if fd != nil {
			err := fd.Close()

			if err != nil {
				return fmt.Errorf("Error closing lock file %s\n", lockPath)
			}

			l.log.Printf("Successfully acquired lock using lockPath : %s\n", lockPath)
			return nil
		}
	}

	return fmt.Errorf("Timed out trying to acquire lock using lockpath %s", lockPath)
}

func (l *FileLock) Unlock() error {

	lockFile := "spectrum-scale-" + l.Filesystem + ".lock"
	lockPath := path.Join(l.Mountpoint, lockFile)

	l.log.Printf("Unlocking on lockPath %s\n", lockPath)

	err := os.Remove(lockPath)

	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Failed to delete lock file %s : %s\n", lockPath, err.Error())
		}
	}

	l.log.Printf("Successfully unlocked on lockpath %s\n", lockPath)
	return nil
}
