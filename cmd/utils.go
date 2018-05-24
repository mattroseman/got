package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

// getGotRootDir searches up the current working directory path for a .got
// directory indicating if the cwd is in a got repository.
// It returns the absolute path to that .got directory
func getGotRootDir() (string, bool) {
	gotRootDir, err := os.Getwd()
	if err != nil {
		return "", false
	}

	for gotRootDir != "/" {
		// check if .got directory is in gotRootDir
		files, err := ioutil.ReadDir(gotRootDir)
		if err != nil {
			return "", false
		}

		for _, file := range files {
			if file.Name() == ".got" {
				return path.Join(gotRootDir, ".got"), true
			}
		}

		// move gotRootDir to parent directory
		gotRootDir = filepath.Dir(gotRootDir)
	}

	return "", false
}
