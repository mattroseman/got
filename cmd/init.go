package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new got repository",
	Long:  "Create a new got repository for managing version control of current directory",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initGot(); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// initGot initializes a new got repository.
// returns error if .got directory already exists
func initGot() error {
	// check if .got directory already exists
	if _, err := os.Stat(".got/"); err == nil {
		return errors.New("This directory is already a got repository")
	}

	// creat .got directory and objects subdirectory
	if err := os.MkdirAll(".got/objects", 0755); err != nil {
		return err
	}

	return nil
}
