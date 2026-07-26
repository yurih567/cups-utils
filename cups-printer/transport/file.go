package transport

import (
	"fmt"
	"os"
)

// File writes raw bytes to a device node or regular file.
type File struct {
	Path string
}

func (f File) String() string {
	return "file:" + f.Path
}

func (f File) Send(payload []byte) error {
	flags := os.O_WRONLY
	fi, err := os.Stat(f.Path)
	switch {
	case os.IsNotExist(err):
		flags |= os.O_CREATE | os.O_TRUNC
	case err != nil:
		return fmt.Errorf("file %s: stat: %w", f.Path, err)
	case fi.Mode().IsRegular():
		flags |= os.O_TRUNC
	}

	file, err := os.OpenFile(f.Path, flags, 0o644)
	if err != nil {
		return fmt.Errorf("file %s: open: %w", f.Path, err)
	}
	defer file.Close()
	if _, err := file.Write(payload); err != nil {
		return fmt.Errorf("file %s: write: %w", f.Path, err)
	}
	return nil
}

// Stdout writes raw bytes to os.Stdout (debug).
type Stdout struct{}

func (Stdout) String() string { return "stdout" }

func (Stdout) Send(payload []byte) error {
	_, err := os.Stdout.Write(payload)
	if err != nil {
		return fmt.Errorf("stdout: write: %w", err)
	}
	return nil
}
