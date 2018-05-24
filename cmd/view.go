package cmd

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view [hash of got object]",
	Short: "View the contents of a got object",
	Long:  "Uncompresses a got object with the given hash, and prints the contents to stdout",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := view(args[0]); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

func view(hash string) error {
	// check that the directory for an object with the given hash exists
	objectFilePath := path.Join(gotRootDir, "objects", hash[:2], hash[2:])
	if _, err := os.Stat(objectFilePath); os.IsNotExist(err) {
		return errors.New("no got object with given hash was found")
	}

	// read in the object file
	objectFile, err := os.Open(objectFilePath)
	if err != nil {
		return err
	}
	defer objectFile.Close()

	// uncompress the object file
	rc, err := zlib.NewReader(objectFile)
	if err != nil {
		return err
	}
	defer rc.Close()

	// split the object file into header and file contents
	var buf bytes.Buffer
	_, err = buf.ReadFrom(rc)
	if err != nil {
		return err
	}
	header, err := buf.ReadBytes('\000')
	if err != nil {
		return err
	}
	content := buf.String()

	fmt.Printf("Header: %s\n", header)
	fmt.Printf("Content:\n%s\n", content)

	return nil
}
