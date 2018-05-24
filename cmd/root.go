package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var gotRootDir string

var rootCmd = &cobra.Command{
	Use:   "got",
	Short: "got is a version control application",
	Long:  "a implementation of git in golang, for learning purposes",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// check to see if current directory is in a got repository
		var ok bool // declare using var because := can't be used with global gotRootDir
		gotRootDir, ok = getGotRootDir()
		if !ok {
			fmt.Println("You are not currently in a got repository")
			os.Exit(1)
		}
	},
}

// Execute starts up cobra, and runs the rootCmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
