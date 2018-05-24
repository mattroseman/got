package cmd

import "github.com/spf13/cobra"

// TODO check that the current directory is a .got repository

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a file to be tracked by .got",
	Long:  "Compresses a file and adds it to .got/objects to be tracked",
	// TODO give this command arguments
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

// add takes a path to a file, and compresses that file adding it to
// .got/objects
func add(filePath string) {
}
