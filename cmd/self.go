/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/tinotenda-alfaneti/labman/internal/remote"
	"github.com/spf13/cobra"
)

// selfCmd groups OS-level maintenance commands for the remote host.
var selfCmd = &cobra.Command{
	Use:   "self",
	Short: "Run maintenance tasks directly on the server",
	Long: `Use the self command for host-level operations such as updating packages,
cleaning disk space, or checking system status without touching the cluster.`,
	PersistentPreRunE: requireSession,
	PersistentPostRun: cleanupSession,
	Run: func(cmd *cobra.Command, args []string) {
		printBanner(cmd)
		printSection(cmd, "SELF", "Run 'labman self info' or other maintenance subcommands")
	},
}

var selfInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show remote server system information",
	Long: `Fetches and displays system information from the remote host,
including OS version, kernel details, and hardware specs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := remote.Current()
		if client == nil {
			return fmt.Errorf("not connected to any server. Run 'labman login' first")
		}

		output, err := client.Run("uname -a && lsb_release -a && lscpu")
		if err != nil {
			return fmt.Errorf("failed to fetch system info: %w", err)
		}

		printSection(cmd, "SYSTEM INFO", output)
		return nil
	},
}

var selfCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Perform safe maintenance and cleanup on the remote server",
	Long: `Runs a series of safe system-level maintenance operations on the remote host:
- Updates and upgrades packages
- Removes old packages and cleans caches
- Vacuums system logs
- Syncs system clock
- Optionally prunes MicroK8s container images`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		client := remote.Current()
		if client == nil {
			return fmt.Errorf("not connected to any server. Run 'labman login' first")
		}

		fmt.Fprintln(out, "===== SYSTEM MAINTENANCE =====")
		fmt.Fprintln(out, "Starting maintenance sequence...")

		// helper to run a command and print status
		run := func(name, command string) error {
			fmt.Fprintf(out, "→ %s\n", name)
			err := client.RunStream(command, out)
			if err != nil {
				fmt.Fprintf(out, "%s failed: %v\n", name, err)
				return err
			}
			fmt.Fprintf(out, "%s completed\n", name)
			return nil
		}

		before, _ := client.Run(`df -h / | awk 'NR==2{print $4}'`)
		before = strings.TrimSpace(before)

		steps := []struct {
			Name    string
			Command string
		}{
			{"Update package lists", "sudo apt-get update -y"},
			{"Upgrade packages", "sudo DEBIAN_FRONTEND=noninteractive apt-get upgrade -yq"},
			{"Clean old packages", "sudo apt-get autoremove -y && sudo apt-get autoclean -y"},
			{"Vacuum logs (7 days)", "sudo journalctl --vacuum-time=7d"},
			{"Sync system clock", "sudo timedatectl set-ntp true"},
			{"Prune container images", "sudo microk8s ctr image prune -a || true"},
		}

		var failed []string
		for i, s := range steps {
			fmt.Fprintf(out, "[%d/%d] %s...\n", i+1, len(steps), s.Name)
			if err := run(s.Name, s.Command); err != nil {
				failed = append(failed, s.Name)
			}
		}

		rebootReq, _ := client.Run("test -f /var/run/reboot-required && echo 'yes' || echo 'no'")
		rebootReq = strings.TrimSpace(rebootReq)
		after, _ := client.Run(`df -h / | awk 'NR==2{print $4}'`)
		after = strings.TrimSpace(after)

		fmt.Fprintln(out, "--------------------------------------------------")
		fmt.Fprintf(out, "Disk free before: %s\n", before)
		fmt.Fprintf(out, "Disk free after : %s\n", after)
		fmt.Fprintf(out, "Reboot required : %s\n", rebootReq)
		if len(failed) > 0 {
			fmt.Fprintf(out, "Failed steps    : %s\n", strings.Join(failed, ", "))
		} else {
			fmt.Fprintln(out, "All steps completed successfully.")
		}
		fmt.Fprintf(out, "Completed at    : %s\n", time.Now().Format(time.RFC1123))
		fmt.Fprintln(out, "--------------------------------------------------")
		return nil
	},
}



func init() {
	rootCmd.AddCommand(selfCmd)
	selfCmd.AddCommand(selfInfoCmd)
	selfCmd.AddCommand(selfCleanCmd)
}
