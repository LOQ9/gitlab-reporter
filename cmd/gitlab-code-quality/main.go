package main

import (
	"gitlab-code-quality/cmd/gitlab-code-quality/commands"
	"os"
)

func main() {
	if err := commands.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
