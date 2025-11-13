# LabMan (Homelab Manager)

LabMan is a Cobra-based CLI that helps you manage a personal homelab from your laptop. It opens SSH sessions to your servers, stores short-lived credentials securely, and provides curated workflows for gathering diagnostics or running maintenance routines.

## Key Features
- **Session-aware commands** `labman login <host>` authenticates over SSH, verifies host keys with your `~/.ssh/known_hosts` file, and caches credentials using the system keyring. `labman session status --drop` lets you audit or rotate cached credentials without hunting for files.
- **Cluster inspection & care** `labman cluster info/status/workloads` tunnel into MicroK8s to show cluster-info dumps, control-plane health, resource usage, and CrashLoop logs. `labman cluster backup` triggers Velero + etcd snapshots, and `labman cluster restart <addon|service>` wraps the usual recovery scripts.
- **Self-maintenance**  `labman self info/clean/disks/services/netcheck/upgrade` cover OS metadata, apt/journal cleanup, disk forensics, critical service checks (with optional restarts), network probes, and Kubernetes-aware OS upgrades.
- **Diagnostic bundles** `labman diag bundle` runs `kubectl get all`, `kubectl get events`, `journalctl`, and `microk8s inspect`, packaging everything into a tarball you can download later or stream directly to stdout.
- **Centralized output helpers** � all commands share consistent banners and boxed sections through `cmd/output.go`, making CLI output easy to scan.

## Prerequisites
- Go 1.24 or newer (per `go.mod`)
- Access to a Linux host reachable over SSH (the maintenance workflow assumes Debian/Ubuntu tooling such as `apt-get`, `journalctl`, `timedatectl`, and MicroK8s)
- A populated SSH `known_hosts` file for your target hosts

## Running Locally
```bash
git clone https://github.com/tinotenda-alfaneti/labman.git
cd labman
go mod tidy            # downloads modules
go build ./...         # optional: compile to ./labman
```

You can also run the CLI without building a binary:

```bash
go run ./main.go login 192.168.1.10 -u ubuntu -p secret
go run ./main.go cluster info
go run ./main.go cluster status
go run ./main.go cluster workloads --max-crash-pods 3
go run ./main.go self clean
go run ./main.go self services --restart microk8s
go run ./main.go diag bundle --stdout > diag.tar.gz
```

The login command must succeed before `cluster` or `self` subcommands run; those rely on the cached session stored under `~/.labman/sessions/credentials.yaml` plus the OS keyring entry (`labman:<user>@<host>`).

## Testing

Once the Go toolchain is installed, run:

```bash
go test ./...
```

Tests focus on the remote-session helpers, ensuring known-host handling, keyring interactions, and session-file parsing work without needing a real SSH server.
