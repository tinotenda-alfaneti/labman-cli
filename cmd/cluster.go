/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tinotenda-alfaneti/labman/internal/remote"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Interact with the current homelab cluster",
	Long: `Cluster hosts all SSH-backed operations. Use it to authenticate,
inspect state, and run other management tasks against your remote server.`,
	Run: func(cmd *cobra.Command, args []string) {
		printBanner(cmd)
		printSection(cmd, "CLUSTER", "Use subcommands such as 'labman cluster info'")
	},
	PersistentPreRunE: requireSession,
	PersistentPostRun: cleanupSession,
}

var clusterInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show Kubernetes cluster diagnostics",
	Long: `Runs 'kubectl cluster-info dump' on the remote host to print detailed cluster
state, making it easy to inspect components from your local terminal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		infoClient := remote.Current()
		if infoClient == nil {
			return fmt.Errorf("not connected to any cluster. Please login first")
		}

		output, err := infoClient.Run("kubectl cluster-info dump")
		if err != nil {
			return fmt.Errorf("failed to run command: %w", err)
		}

		printSection(cmd, "CLUSTER INFO", output)

		return nil
	},
}

var clusterStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Summarize MicroK8s readiness and control-plane health",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := remote.Current()
		if client == nil {
			return fmt.Errorf("not connected to any cluster. Please login first")
		}

		steps := []struct {
			Title   string
			Command string
		}{
			{"MICROK8S STATUS", "microk8s status --wait-ready"},
			{"NODES", "microk8s kubectl get nodes -o wide"},
			{"COMPONENT HEALTH", "microk8s kubectl get componentstatuses"},
		}

		for _, step := range steps {
			output, err := client.Run(step.Command)
			if err != nil {
				return fmt.Errorf("%s failed: %w", strings.ToLower(step.Title), err)
			}
			printSection(cmd, step.Title, output)
		}
		return nil
	},
}

var clusterWorkloadsCmd = &cobra.Command{
	Use:   "workloads",
	Short: "Inspect pods, resource usage, and CrashLoop logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace, _ := cmd.Flags().GetString("namespace")
		selector, _ := cmd.Flags().GetString("selector")
		maxPods, _ := cmd.Flags().GetInt("max-crash-pods")
		if maxPods <= 0 {
			maxPods = 5
		}
		logTail, _ := cmd.Flags().GetInt("logs-tail")
		if logTail <= 0 {
			logTail = 20
		}

		client := remote.Current()
		if client == nil {
			return fmt.Errorf("not connected to any cluster. Please login first")
		}

		nsArg := namespaceArg(namespace)
		selectorArg := selectorArg(selector)

		nonRunningCmd := joinCommand(
			"microk8s kubectl get pods",
			nsArg,
			selectorArg,
			"--field-selector=status.phase!=Running,status.phase!=Succeeded",
			"--sort-by=.status.containerStatuses[0].restartCount",
		)
		nonRunning, err := client.Run(nonRunningCmd)
		if err != nil {
			return fmt.Errorf("list non-running pods: %w", err)
		}
		printSection(cmd, "NON-RUNNING PODS", nonRunning)

		topCmd := joinCommand("microk8s kubectl top pods", nsArg, selectorArg)
		topOutput, err := client.Run(topCmd)
		if err != nil {
			topOutput = fmt.Sprintf("kubectl top pods failed: %v", err)
		}
		printSection(cmd, "POD RESOURCE USAGE", topOutput)

		crashPods, err := fetchCrashLoopPods(client, nsArg, selectorArg, maxPods)
		if err != nil {
			printSection(cmd, "CRASHLOOP PODS", fmt.Sprintf("failed to discover CrashLoopBackOff pods: %v", err))
			return nil
		}

		if len(crashPods) == 0 {
			printSection(cmd, "CRASHLOOP PODS", "No pods currently in CrashLoopBackOff.")
			return nil
		}

		var list strings.Builder
		for _, pod := range crashPods {
			fmt.Fprintf(&list, "%s/%s\n", pod.Namespace, pod.Name)
		}
		printSection(cmd, "CRASHLOOP PODS", strings.TrimSpace(list.String()))

		out := cmd.OutOrStdout()
		fmt.Fprintln(out, "===== CRASHLOOP LOGS =====")
		for _, pod := range crashPods {
			fmt.Fprintf(out, "--- %s/%s ---\n", pod.Namespace, pod.Name)
			logCmd := joinCommand(
				"microk8s kubectl logs",
				"-n "+shellQuote(pod.Namespace),
				shellQuote(pod.Name),
				"--all-containers",
				fmt.Sprintf("--tail=%d", logTail),
			)
			if err := client.RunStream(logCmd, out); err != nil {
				fmt.Fprintf(out, "failed to fetch logs: %v\n", err)
			}
			fmt.Fprintln(out)
		}

		return nil
	},
}

var clusterBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create or list Velero/etcd snapshots",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := remote.Current()
		if client == nil {
			return fmt.Errorf("not connected to any cluster. Please login first")
		}

		listOnly, _ := cmd.Flags().GetBool("list")
		name, _ := cmd.Flags().GetString("name")
		skipVelero, _ := cmd.Flags().GetBool("skip-velero")
		skipEtcd, _ := cmd.Flags().GetBool("skip-etcd")

		if listOnly {
			return showBackupInventory(cmd, client)
		}

		if skipVelero && skipEtcd {
			return fmt.Errorf("nothing to do: both Velero and etcd backups are disabled")
		}

		if name == "" {
			name = fmt.Sprintf("labman-%s", time.Now().Format("20060102-150405"))
		}

		if !skipVelero {
			veleroCmd := fmt.Sprintf("microk8s velero create backup %s --ttl 720h", shellQuote(name))
			output, err := client.Run(veleroCmd)
			if err != nil {
				return fmt.Errorf("velero backup failed: %w", err)
			}
			printSection(cmd, "VELERO BACKUP", output)
		}

		if !skipEtcd {
			etcdPath := fmt.Sprintf("/var/snap/microk8s/common/var/backup/labman-etcd-%s.db", time.Now().Format("20060102-150405"))
			etcdCmd := fmt.Sprintf("sudo mkdir -p /var/snap/microk8s/common/var/backup && sudo microk8s etcd snapshot save %s", shellQuote(etcdPath))
			output, err := client.Run(etcdCmd)
			if err != nil {
				return fmt.Errorf("etcd snapshot failed: %w", err)
			}
			printSection(cmd, "ETCD SNAPSHOT", output+"\nSaved to: "+etcdPath)
		}

		return nil
	},
}

var clusterRestartCmd = &cobra.Command{
	Use:   "restart <target>",
	Short: "Restart a MicroK8s addon or snap service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		mode, _ := cmd.Flags().GetString("type")
		waitReady, _ := cmd.Flags().GetBool("wait")

		client := remote.Current()
		if client == nil {
			return fmt.Errorf("not connected to any cluster. Please login first")
		}

		var restartCmd string
		switch mode {
		case "addon":
			restartCmd = strings.Join([]string{
				"sudo microk8s disable " + shellQuote(target),
				"sudo microk8s enable " + shellQuote(target),
			}, " && ")
		case "service":
			restartCmd = "sudo systemctl restart " + shellQuote(target)
		default:
			return fmt.Errorf("unknown restart type %q (use 'addon' or 'service')", mode)
		}

		if _, err := client.Run(restartCmd); err != nil {
			return fmt.Errorf("restart failed: %w", err)
		}

		printSection(cmd, "RESTART", fmt.Sprintf("Successfully restarted %s (%s)", target, mode))

		if waitReady && mode == "addon" {
			status, err := client.Run("microk8s status --wait-ready")
			if err != nil {
				return fmt.Errorf("microk8s did not become ready: %w", err)
			}
			printSection(cmd, "MICROK8S STATUS", status)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
	clusterCmd.AddCommand(clusterInfoCmd)
	clusterCmd.AddCommand(clusterStatusCmd)
	clusterCmd.AddCommand(clusterWorkloadsCmd)
	clusterCmd.AddCommand(clusterBackupCmd)
	clusterCmd.AddCommand(clusterRestartCmd)

	clusterWorkloadsCmd.Flags().StringP("namespace", "n", "", "namespace to scope workload checks (default: all)")
	clusterWorkloadsCmd.Flags().StringP("selector", "l", "", "label selector to filter pods (e.g. app=web)")
	clusterWorkloadsCmd.Flags().Int("max-crash-pods", 5, "maximum number of CrashLoopBackOff pods to inspect")
	clusterWorkloadsCmd.Flags().Int("logs-tail", 20, "number of log lines to fetch for each CrashLoopBackOff pod")

	clusterBackupCmd.Flags().Bool("list", false, "list existing backups instead of creating new ones")
	clusterBackupCmd.Flags().String("name", "", "custom name for the Velero backup (default: labman-<timestamp>)")
	clusterBackupCmd.Flags().Bool("skip-velero", false, "skip Velero backup creation")
	clusterBackupCmd.Flags().Bool("skip-etcd", false, "skip etcd snapshot creation")

	clusterRestartCmd.Flags().String("type", "addon", "what to restart: addon or service")
	clusterRestartCmd.Flags().Bool("wait", true, "wait for microk8s to report ready after restarting an addon")
}

func namespaceArg(namespace string) string {
	if namespace == "" {
		return "-A"
	}
	return "-n " + shellQuote(namespace)
}

func selectorArg(selector string) string {
	if selector == "" {
		return ""
	}
	return "--selector=" + shellQuote(selector)
}

type podRef struct {
	Namespace string
	Name      string
}

type kubernetesPodList struct {
	Items []struct {
		Metadata struct {
			Namespace string `json:"namespace"`
			Name      string `json:"name"`
		} `json:"metadata"`
		Status struct {
			ContainerStatuses []struct {
				State struct {
					Waiting *struct {
						Reason string `json:"reason"`
					} `json:"waiting"`
				} `json:"state"`
			} `json:"containerStatuses"`
		} `json:"status"`
	} `json:"items"`
}

func fetchCrashLoopPods(client *remote.SSHSession, nsArg, selectorArg string, limit int) ([]podRef, error) {
	jsonCmd := joinCommand("microk8s kubectl get pods", nsArg, selectorArg, "-o json") + " 2>/dev/null"
	raw, err := client.Run(jsonCmd)
	if err != nil {
		return nil, err
	}

	var list kubernetesPodList
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil, fmt.Errorf("parse kubectl json: %w", err)
	}

	pods := make([]podRef, 0)
	for _, item := range list.Items {
		for _, cs := range item.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
				pods = append(pods, podRef{Namespace: item.Metadata.Namespace, Name: item.Metadata.Name})
				break
			}
		}
		if limit > 0 && len(pods) >= limit {
			break
		}
	}

	return pods, nil
}

func showBackupInventory(cmd *cobra.Command, client *remote.SSHSession) error {
	velero, err := client.Run("microk8s velero backup get")
	if err != nil {
		velero = fmt.Sprintf("velero backup listing failed: %v", err)
	}
	printSection(cmd, "VELERO BACKUPS", velero)

	etcd, err := client.Run("sudo microk8s etcd snapshot list")
	if err != nil {
		etcd = fmt.Sprintf("etcd snapshot listing failed: %v", err)
	}
	printSection(cmd, "ETCD SNAPSHOTS", etcd)
	return nil
}
