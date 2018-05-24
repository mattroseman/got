package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "got",
	Short: "got is a version control application",
	Long:  "a implementation of git in golang, for learning purposes",
}

// Execute starts up cobra, and runs the rootCmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
