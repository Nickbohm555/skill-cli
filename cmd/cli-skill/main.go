package main

import (
	"fmt"
	"os"

	"github.com/Nickbohm555/skill-cli/internal/cli/command"
)

func main() {
	rootCmd := command.NewRootCommand()
	rootCmd.SetErr(os.Stderr)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
