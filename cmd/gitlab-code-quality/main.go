package main

import (
	"os"

	"github.com/LOQ9/gitlab-code-quality/cmd/gitlab-code-quality/commands"
)

func main() {
	if err := commands.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
