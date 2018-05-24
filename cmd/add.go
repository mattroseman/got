package cmd

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
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
	fmt.Printf("SHA-1 hash: %s\n", hash)

	// zlib compress file contents
	wc := zlib.NewWriter(os.Stdout)
	defer wc.Close()

	fmt.Printf("zlib compression: ")
	wc.Write(buf.Bytes())
	fmt.Println()

	// TODO store compressed file in .got/objects directory with dir/filenam matching hash

	return nil
}
