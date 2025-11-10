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

var selfUpgradeOSCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Safely upgrade the OS on the current MicroK8s node without disrupting workloads",
	Long: `Performs a safe, Kubernetes-aware OS upgrade on the current node:

1. Detects Pi-hole and warns
2. Cordon + drain the node (move workloads away)
3. Run non-interactive apt upgrade
4. Optionally refresh microk8s
5. Reboot
6. Wait for node + microk8s to be ready again
7. Uncordon the node

This command assumes you have a working SSH session via 'labman login'.`,
	PreRunE: requireSession, // your existing helper
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		client := remote.Current()
		if client == nil {
			return fmt.Errorf("not connected to any server. Run 'labman login' first")
		}

		doRefreshMicrok8s, _ := cmd.Flags().GetBool("refresh-microk8s")
		skipReboot, _ := cmd.Flags().GetBool("no-reboot")

		fmt.Fprintln(out, "===== CLUSTER OS UPGRADE =====")

		hostnameRaw, err := client.Run("hostname")
		if err != nil {
			return fmt.Errorf("failed to detect hostname: %w", err)
		}
		nodeName := strings.TrimSpace(hostnameRaw)
		fmt.Fprintf(out, "Target node: %s\n", nodeName)

		fmt.Fprintln(out, "→ Checking for Pi-hole on host ...")
		piholeStatus, _ := client.Run("systemctl is-active pihole-FTL 2>/dev/null || true")
		if strings.TrimSpace(piholeStatus) == "active" {
			fmt.Fprintln(out, "Pi-hole is running directly on this host.")
			fmt.Fprintln(out, "DNS may be unavailable while this node reboots.")
		}

		// 3) check for velero and attempt pre-upgrade backup (best-effort)
		fmt.Fprintln(out, "→ Checking for Velero ...")
		_, _ = client.Run("microk8s kubectl get ns velero >/dev/null 2>&1 || true")
		_, _ = client.Run(`microk8s kubectl get ns velero >/dev/null 2>&1 && microk8s velero create backup pre-upgrade-$(date +%F-%H%M) || true`)
		fmt.Fprintln(out, "Velero check complete (backup best-effort).")

		fmt.Fprintf(out, "→ Cordoning node %s ...\n", nodeName)
		if _, err := client.Run("sudo microk8s kubectl cordon " + nodeName); err != nil {
			return fmt.Errorf("failed to cordon node: %w", err)
		}
		fmt.Fprintln(out, "Node cordoned.")

		fmt.Fprintf(out, "→ Draining node %s ...\n", nodeName)
		drainCmd := "sudo microk8s kubectl drain " + nodeName + " --ignore-daemonsets --delete-emptydir-data --force"
		if _, err := client.Run(drainCmd); err != nil {
			fmt.Fprintf(out, "drain reported an error: %v\n", err)
		} else {
			fmt.Fprintln(out, "Node drained.")
		}

		fmt.Fprintln(out, "→ Updating package lists & upgrading OS (non-interactive) ...")
		upgradeCmd := "sudo apt-get update -y && sudo DEBIAN_FRONTEND=noninteractive apt-get full-upgrade -yq && sudo do-release-upgrade"
		if _, err := client.Run(upgradeCmd); err != nil {
			return fmt.Errorf("failed to run OS upgrade: %w", err)
		}
		fmt.Fprintln(out, "OS packages upgraded.")

		if doRefreshMicrok8s {
			fmt.Fprintln(out, "→ Refreshing microk8s snap ...")
			if _, err := client.Run("sudo snap refresh microk8s"); err != nil {
				fmt.Fprintf(out, "microk8s refresh failed: %v\n", err)
			} else {
				fmt.Fprintln(out, "microk8s refreshed.")
			}
		}

		if skipReboot {
			fmt.Fprintln(out, "→ Skipping reboot as requested (--no-reboot).")
		} else {
			fmt.Fprintln(out, "→ Rebooting node ...")
			// fire-and-forget reboot; SSH will drop
			_, _ = client.Run("sudo reboot now || sudo shutdown -r now || true")
		}

		if !skipReboot {
			fmt.Fprintln(out, "→ Waiting for node to return (this may take a few minutes)...")

			var ready bool
			for i := 0; i < 30; i++ {
				time.Sleep(20 * time.Second)

				newClient, err := remote.LoadSession()
				if err != nil || newClient == nil {
					fmt.Fprintln(out, "... node not up yet (SSH not ready)")
					continue
				}

				status, err := newClient.Run("microk8s status --wait-ready")
				if err != nil {
					fmt.Fprintln(out, "... microk8s not ready yet")
					continue
				}
				if strings.Contains(status, "microk8s is running") {
					fmt.Fprintln(out, "Node and microk8s are back.")
					remote.SetCurrent(newClient)
					ready = true
					break
				}
			}

			if !ready {
				fmt.Fprintln(out, "Node did not come back in time. Please check manually.")
				fmt.Fprintf(out, "You may need to run: microk8s kubectl uncordon %s\n", nodeName)
				return nil
			}
		}

		fmt.Fprintf(out, "→ Uncordoning node %s ...\n", nodeName)
		newClient := remote.Current()
		if newClient == nil {
			newClient, _ = remote.LoadSession()
		}
		if newClient != nil {
			if _, err := newClient.Run("sudo microk8s kubectl uncordon " + nodeName); err != nil {
				fmt.Fprintf(out, "failed to uncordon node: %v\n", err)
			} else {
				fmt.Fprintln(out, "Node uncordoned.")
			}
		} else {
			fmt.Fprintln(out, "could not restore SSH session to uncordon; please uncordon manually.")
		}

		fmt.Fprintln(out, "--------------------------------------------------")
		fmt.Fprintln(out, "OS upgrade flow completed.")
		fmt.Fprintf(out, "Node: %s\n", nodeName)
		if skipReboot {
			fmt.Fprintln(out, "Note: reboot was skipped, you should reboot this node to finish the upgrade.")
		}
		fmt.Fprintln(out, "--------------------------------------------------")

		return nil
	},
}


func init() {
	rootCmd.AddCommand(selfCmd)
	selfCmd.AddCommand(selfInfoCmd)
	selfCmd.AddCommand(selfCleanCmd)
	selfCmd.AddCommand(selfUpgradeOSCmd)
	selfUpgradeOSCmd.Flags().Bool("refresh-microk8s", false, "refresh the microk8s snap after OS upgrade")
	selfUpgradeOSCmd.Flags().Bool("no-reboot", false, "perform the upgrade but do not reboot (for testing)")
}
