package utils

import "os"

func GetHome() string {
	winHome := os.Getenv("userprofile")
	unixHome := os.Getenv("HOME")
	if winHome != "" {
		return winHome
	}

	if unixHome != "" {
		return unixHome
	}
	panic("unknown Operating System")
}
