package main

import (
	"os"

	"github.com/LOQ9/gitlab-reporter/cmd/gitlab-reporter/commands"
)

func main() {
	if err := commands.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
