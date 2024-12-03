package stores

import (
	"io"
	"testing"

	"github.com/AnthonyDickson/yatta/yattatest"
)

func TestTapeWrite(t *testing.T) {
	t.Run("write shorter data", func(t *testing.T) {
		file, cleanup := yattatest.CreateTempFile(t, "12345")
		defer cleanup()

		tp := tape{file}

		_, err := tp.Write([]byte("abc"))
		yattatest.AssertNoError(t, err)

		_, err = file.Seek(0, io.SeekStart)
		yattatest.AssertNoError(t, err)

		fileContents, err := io.ReadAll(file)
		yattatest.AssertNoError(t, err)

		got := string(fileContents)
		want := "abc"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
}
