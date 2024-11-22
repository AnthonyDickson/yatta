package yattatest

import (
	"os"
	"testing"
)

// Creates a temporary file for testing.
//
// Returns the file and a function for removing the file.
func CreateTempFile(t *testing.T, initialData string) (*os.File, func()) {
	t.Helper()

	tempFile, err := os.CreateTemp("", "db")

	if err != nil {
		t.Fatalf("could not create temporary file: %v", err)
	}

	_, err = tempFile.Write([]byte(initialData))

	if err != nil {
		t.Fatalf("could not write initial data to temp file: %v", err)
	}

	removeFile := func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}

	return tempFile, removeFile
}
