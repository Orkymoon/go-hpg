package helper

import "time"

func IsEmpty(s string) bool {
	return len(s) == 0
}

func IsErrorMessage(err error, m string) bool {
	return err.Error() == m
}

func UnixToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func Contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
