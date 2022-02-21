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
	Use:   "gitlab-code-quality",
	Short: "Gitlab Code Quality",
	Long:  `Code Quality parser to compatible Gitlab report`,
}
