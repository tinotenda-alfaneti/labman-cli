package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const homelabBanner = `Welcome to LabMan - Your Homelab Management CLI`

func printBanner(cmd *cobra.Command) { fmt.Fprintln(cmd.OutOrStdout(), homelabBanner)}

func printSection(cmd *cobra.Command, title, body string) {
	out := cmd.OutOrStdout()
	bodyLines := splitLines(body)
	content := append([]string{title, ""}, bodyLines...)

	maxWidth := 0
	for _, line := range content {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	border := "+" + strings.Repeat("-", maxWidth+2) + "+"
	fmt.Fprintln(out, border)
	for _, line := range content {
		padding := strings.Repeat(" ", maxWidth-len(line))
		fmt.Fprintf(out, "| %s%s |\n", line, padding)
	}
	fmt.Fprintln(out, border)
}

func splitLines(body string) []string {
	body = strings.TrimRight(body, "\n")
	if body == "" {
		return []string{""}
	}
	return strings.Split(body, "\n")
}
