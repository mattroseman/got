package cmd

import "github.com/spf13/cobra"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new got repository",
	Long:  "Create a new got repository for managing version control of current directory",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO create .got directory
		// TODO create .got/objects directory
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
