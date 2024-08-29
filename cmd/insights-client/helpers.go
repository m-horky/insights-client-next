package main

import (
	"os"
	"time"
)

// writeTimestampFile saves current timestamp into a file.
func writeTimestampFile(path string) error {
	now := time.Now()
	timestamp := now.Format("2006-01-02T15:04:05.999Z07:00")
	return os.WriteFile(path, []byte(timestamp), 0775)
}
