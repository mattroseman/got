package cmd

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [path to file to add]",
	Short: "Add a file to be tracked by .got",
	Long:  "Compresses a file and adds it to .got/objects to be tracked",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := add(args[0]); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

// add takes a path to a file, and compresses that file adding it to
// .got/objects
func add(filePath string) error {
	// get absolute path to current working directory
	workingDirPath, err := os.Getwd()
	if err != nil {
		return err
	}

	// get absolute filepath for given filePath
	fileAbsPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	// check that filePath is within the current directory or subdirectories
	if !strings.HasPrefix(fileAbsPath, workingDirPath) {
		return errors.New("file is not in current got repository")
	}

	// check if file exists
	if _, err := os.Stat(fileAbsPath); os.IsNotExist(err) {
		return fmt.Errorf("no file at %s was found", filePath)
	}

	// read in file at given path
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bytes.NewBuffer(make([]byte, 0))
	tee := io.TeeReader(f, buf)

	// SHA-1 hash file contents
	h := sha1.New()
	if _, err := io.Copy(h, tee); err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))

	// make object directory named after first 2 bytes of sha-1 hash if it doesn't already exist
	objectDir := path.Join(gotRootDir, "objects", hash[:2])
	if _, err := os.Stat(objectDir); os.IsNotExist(err) {
		if err := os.Mkdir(objectDir, 0755); err != nil {
			return err
		}
	}

	// TODO what if this file has already been added
	// create file for storing the compressed contents of this file
	objectFileName := hash[2:]
	objectFile, err := os.Create(path.Join(objectDir, objectFileName))
	if err != nil {
		return err
	}
	defer objectFile.Close()

	// zlib compress file contents and store in objects directory
	wc := zlib.NewWriter(objectFile)
	wc.Write(buf.Bytes())
	wc.Close()

	return nil
}
