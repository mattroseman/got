package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new got repository",
	Long:  "Create a new got repository for managing version control of current directory",
	Args:  cobra.NoArgs,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// check if .got directory already exists
		if _, ok := getGotRootDir(); ok {
			fmt.Println("You are already in a got repository")
			os.Exit(1)
		}
	},
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
	// creat .got directory and objects subdirectory
	if err := os.MkdirAll(".got/objects", 0755); err != nil {
		return err
	}

	// TODO rename this later
	// create empty master tree
	// fileBytes := []byte("tree 0\000")

	return nil
}
