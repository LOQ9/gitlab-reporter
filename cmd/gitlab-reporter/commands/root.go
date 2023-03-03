package commands

import (
	"github.com/spf13/cobra"
)

// Run ...
func Run(args []string) error {
	RootCmd.SetArgs(args)
	return RootCmd.Execute()
}

// RootCmd ..
var RootCmd = &cobra.Command{
	Use:   "gitlab-reporter",
	Short: "Gitlab Reporter",
	Long:  `Swiss army knife reporter tool for Gitlab`,
}
