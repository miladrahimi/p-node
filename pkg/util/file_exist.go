package util

import "os"

// FileExist checks if the given file path exists or not.
func FileExist(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}
