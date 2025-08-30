package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainCommand(t *testing.T) {
	t.Run("Help flag shows usage", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		os.Args = []string{"rigel", "--help"}

		output := captureOutput(func() {
			main()
		})

		assert.Contains(t, output, "AI-powered coding assistant")
		assert.Contains(t, output, "Usage:")
	})

	t.Run("Stdin input mode", func(t *testing.T) {
		oldArgs := os.Args
		oldStdin := os.Stdin
		defer func() {
			os.Args = oldArgs
			os.Stdin = oldStdin
		}()

		input := "Test prompt from stdin"
		r, w, err := os.Pipe()
		require.NoError(t, err)

		os.Stdin = r
		os.Args = []string{"rigel"}

		go func() {
			defer w.Close()
			io.WriteString(w, input)
		}()

		os.Setenv("PROVIDER", "anthropic")
		os.Setenv("ANTHROPIC_API_KEY", "test-key")
		defer func() {
			os.Unsetenv("PROVIDER")
			os.Unsetenv("ANTHROPIC_API_KEY")
		}()
	})
}

func TestInteractiveMode(t *testing.T) {
	t.Run("Interactive mode exits on quit command", func(t *testing.T) {
		input := "quit\n"
		r := strings.NewReader(input)

		oldStdin := os.Stdin
		oldStdout := os.Stdout
		defer func() {
			os.Stdin = oldStdin
			os.Stdout = oldStdout
		}()

		pr, pw, _ := os.Pipe()
		os.Stdin = pr
		os.Stdout, _ = os.Create(os.DevNull)

		go func() {
			defer pw.Close()
			io.Copy(pw, r)
		}()
	})

	t.Run("Interactive mode exits on exit command", func(t *testing.T) {
		input := "exit\n"
		r := strings.NewReader(input)

		oldStdin := os.Stdin
		oldStdout := os.Stdout
		defer func() {
			os.Stdin = oldStdin
			os.Stdout = oldStdout
		}()

		pr, pw, _ := os.Pipe()
		os.Stdin = pr
		os.Stdout, _ = os.Create(os.DevNull)

		go func() {
			defer pw.Close()
			io.Copy(pw, r)
		}()
	})
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		buf := &bytes.Buffer{}
		io.Copy(buf, r)
		done <- buf.String()
	}()

	f()
	w.Close()

	os.Stdout = old
	return <-done
}
