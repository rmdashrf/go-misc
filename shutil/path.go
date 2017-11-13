package shutil

import "os"

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	return !os.IsNotExist(err), err
}

func PathIsDir(path string) (bool, error) {
	st, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return st.IsDir(), nil
}

func PathIsFile(path string) (bool, error) {
	st, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return !st.IsDir(), nil
}
