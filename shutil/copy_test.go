package shutil

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestMoveCrossDomain(t *testing.T) {
	var (
		err error
	)
	homeDir := os.Getenv("OTHER_DRIVE")
	if homeDir == "" {
		t.Skipf("OTHER_DRIVE env not set")
	}
	tmpDir, err := ioutil.TempDir("", "shutil-test")
	if err != nil {
		t.Fatalf("could not create temporary directory. what kind of shitty environment are you running?")
	}
	defer os.RemoveAll(tmpDir)

	tmpDir2, err := ioutil.TempDir(homeDir, "shutil-test")
	if err != nil {
		t.Fatalf("could not create secondary temporary directory.")
	}
	defer os.RemoveAll(tmpDir2)

	tmpPath := path.Join(tmpDir, "herp")
	finalPath := path.Join(tmpDir2, "derp")

	f, err := os.Create(tmpPath)
	if err != nil {
		t.Fatalf("couldnt create file: %v", err)
	}
	f.WriteString("derp")
	f.Close()

	t.Logf("moving %s to %s", tmpPath, finalPath)

	if err := MoveFile(tmpPath, finalPath); err != nil {
		t.Errorf("could not move file: %v", err)
	}
}
