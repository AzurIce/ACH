package utils

import "time"

func GetTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func FormatTime(time time.Time) string {
	return time.Format("2006-01-02 15:04:05")
}
