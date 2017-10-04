package shutil

import (
	"io"
	"os"
	"syscall"
)

// Copies a file from src to dst.
func CopyFile(src, dst string) error {
	var (
		f   *os.File
		g   *os.File
		err error
	)

	if f, err = os.Open(src); err != nil {
		return err
	}

	defer f.Close()

	if g, err = os.Create(dst); err != nil {
		return err
	}
	defer g.Close()

	_, err = io.Copy(g, f)
	return err
}

// Moves the file. Not guaranteed to be atomic.
func MoveFile(src, dst string) (err error) {
	if err = os.Rename(src, dst); err == nil {
		return nil
	}

	// Copy the file if it is a cross device move then delete the original file
	if isCrossDeviceMove(err) {
		if err := CopyFile(src, dst); err != nil {
			return err
		}

		if err := os.Remove(src); err != nil {
			return err
		}
	}

	return nil
}

func isCrossDeviceMove(err error) bool {
	linkError, ok := err.(*os.LinkError)
	if !ok {
		return false
	}

	errno, ok := linkError.Err.(syscall.Errno)
	if !ok {
		return false
	}

	return errno == syscall.EXDEV
}
