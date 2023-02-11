package util

import (
	"log"
	"os"
)

// CompareFileContents reports whether contents of two files are the same or not
func CompareFileContents(file1, file2 string) (bool, error) {
	data1, err := os.ReadFile(file1)
	if err != nil {
		log.Printf("error reading %s: %s", file1, err)
		return false, err
	}

	data2, err := os.ReadFile(file2)
	if err != nil {
		log.Printf("error reading %s: %s", file2, err)
		return false, err
	}

	return string(data1) == string(data2), nil
}
