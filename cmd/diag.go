package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tinotenda-alfaneti/labman/internal/remote"
)

var diagCmd = &cobra.Command{
	Use:   "diag",
	Short: "Collect diagnostic bundles for troubleshooting",
	Long: `The diag commands gather Kubernetes manifests, MicroK8s logs,
and journal output into a single tarball you can download or stream.`,
	PersistentPreRunE: requireSession,
	PersistentPostRun: cleanupSession,
	Run: func(cmd *cobra.Command, args []string) {
		printBanner(cmd)
		printSection(cmd, "DIAG", "Use 'labman diag bundle' to collect support artifacts.")
	},
}

var diagBundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Create a tarball with kubectl output and MicroK8s diagnostics",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := remote.Current()
		if client == nil {
			return fmt.Errorf("not connected to any cluster. Please login first")
		}

		remotePath, _ := cmd.Flags().GetString("remote-path")
		if remotePath == "" {
			remotePath = fmt.Sprintf("/tmp/labman-diag-%s.tar.gz", time.Now().Format("20060102-150405"))
		}
		stream, _ := cmd.Flags().GetBool("stdout")

		script := buildDiagBundleScript(remotePath)
		result, err := client.Run(script)
		if err != nil {
			return fmt.Errorf("failed to build diagnostic bundle: %w", err)
		}

		path := strings.TrimSpace(result)
		summary := fmt.Sprintf(`Created bundle at %s
Contents:
- kubectl get all -A -o yaml
- kubectl get events -A --sort-by=.lastTimestamp
- kubectl describe nodes
- journalctl -u 'snap.microk8s*' --since -2h
- microk8s inspect output`, path)

		if stream {
			fmt.Fprintf(cmd.ErrOrStderr(), "Streaming diagnostic bundle from %s ...\n", path)
			if err := client.RunStream("cat "+shellQuote(path), cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("stream bundle: %w", err)
			}
			return nil
		}

		printSection(cmd, "DIAGNOSTIC BUNDLE", summary)
		fmt.Fprintf(cmd.OutOrStdout(), "\nRetrieve it later with: scp <host>:%s ./\n", path)
		return nil
	},
}

func buildDiagBundleScript(remotePath string) string {
	return fmt.Sprintf(`set -euo pipefail
TMP_DIR=$(mktemp -d /tmp/labman-diag-XXXXXX)
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT
microk8s kubectl get all -A -o yaml > "$TMP_DIR/k8s-all.yaml"
microk8s kubectl get events -A --sort-by=.lastTimestamp > "$TMP_DIR/k8s-events.txt" || true
microk8s kubectl describe nodes > "$TMP_DIR/k8s-nodes.txt" || true
sudo journalctl -u 'snap.microk8s*' --since -2h > "$TMP_DIR/microk8s-journal.txt" || true
microk8s inspect > "$TMP_DIR/microk8s-inspect.txt" || true
mkdir -p "$(dirname %s)"
tar -czf %s -C "$TMP_DIR" .
echo %s`, shellQuote(remotePath), shellQuote(remotePath), shellQuote(remotePath))
}

func init() {
	rootCmd.AddCommand(diagCmd)
	diagCmd.AddCommand(diagBundleCmd)

	diagBundleCmd.Flags().String("remote-path", "", "Where to store the bundle on the server (default: /tmp/labman-diag-<timestamp>.tar.gz)")
	diagBundleCmd.Flags().Bool("stdout", false, "Stream the tarball to stdout (use with shell redirection)")
}
