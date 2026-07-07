package csv

import (
	"os"

	"github.com/gocarina/gocsv"
)

func ReadFile(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return ReadBytes(content, target)
}

func ReadBytes(content []byte, target any) error {
	return gocsv.UnmarshalBytes(content, target)
}
