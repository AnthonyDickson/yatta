// Package yattatest implements helper functions for unit tests.
package yattatest

import "testing"

func AssertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("got an error %v, want no error", err)
	}
}
