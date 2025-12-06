package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "payments",
	}

	cmd.AddCommand(NewServeCmd())

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	rootCmd := NewRootCmd()
	err := rootCmd.Execute()
	if err != nil {
		return 1
	}
	return 0
}
