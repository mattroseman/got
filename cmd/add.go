package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TODO check that the current directory is a .got repository

var addCmd = &cobra.Command{
	Use:   "add [path to file to add]",
	Short: "Add a file to be tracked by .got",
	Long:  "Compresses a file and adds it to .got/objects to be tracked",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		add(args[0])
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

// add takes a path to a file, and compresses that file adding it to
// .got/objects
func add(filePath string) {
	fmt.Printf("TODO add %s to .got directory\n", filePath)
}
