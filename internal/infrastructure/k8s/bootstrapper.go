package k8s

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/platform/buildinfo"
	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

//go:embed manifests/*.yaml
var manifestFS embed.FS

const (
	cmdHelm             = "helm"
	cmdRepo             = "repo"
	cmdAdd              = "add"
	cmdUpdate           = "update"
	cmdUpgrade          = "upgrade"
	flagInstall         = "--install"
	flagNamespace       = "--namespace"
	flagCreateNamespace = "--create-namespace"
	nsMonitoring        = "monitoring"
	nsArgoCD            = "argocd"
)

// CommandExecutor runs a CLI command (helm/kubectl) with an optional scoped kubeconfig path.
type CommandExecutor func(ctx context.Context, kubeconfig string, name string, args ...string) ([]byte, error)

// Bootstrapper implements ports.ClusterBootstrapper for local and remote K3s clusters.
type Bootstrapper struct {
	execCmd      CommandExecutor
	credStore    ports.CredentialStore
	agentVersion string
}

// NewBootstrapper creates a Kubernetes Bootstrapper with command-scoped KUBECONFIG execution.
func NewBootstrapper(credStore ports.CredentialStore) *Bootstrapper {
	return &Bootstrapper{
		execCmd:      defaultExecutor,
		credStore:    credStore,
		agentVersion: buildinfo.Version,
	}
}

// NewBootstrapperWithExecutor creates a Bootstrapper with a custom CommandExecutor for unit tests.
func NewBootstrapperWithExecutor(execFn CommandExecutor) *Bootstrapper {
	return &Bootstrapper{
		execCmd:      execFn,
		agentVersion: "dev",
	}
}

func defaultExecutor(ctx context.Context, kubeconfig string, name string, args ...string) ([]byte, error) {
	//nolint:gosec
	cmd := exec.CommandContext(ctx, name, args...)
	if kubeconfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
	}
	return cmd.CombinedOutput()
}

// BootstrapCluster provisions K3s over SSH when a remote host is specified, then installs requested Helm add-ons.
//
//nolint:gocyclo
func (b *Bootstrapper) BootstrapCluster(ctx context.Context, opts ports.ClusterBootstrapOptions) (ports.ClusterState, error) {
	kubeconfigPath := ""
	domain := opts.Domain

	if opts.Provider == catalog.ClusterAWSEC2 && opts.Host != "" {
		var err error
		kubeconfigPath, err = b.bootstrapRemoteK3s(ctx, opts)
		if err != nil {
			return ports.ClusterState{}, err
		}
		if domain == "" {
			domain = opts.Host + ".nip.io"
		}
	}
	if domain == "" {
		domain = "127.0.0.1.nip.io"
	}

	if opts.InstallArgoCD {
		progress.Emit(ctx, "cluster-argocd", progress.StatusRunning, "installing ArgoCD via Helm")
		if err := b.installArgoCD(ctx, kubeconfigPath, opts.Provider, domain); err != nil {
			return ports.ClusterState{}, err
		}
		progress.Emit(ctx, "cluster-argocd", progress.StatusDone, "ArgoCD installed in namespace argocd")
	}

	if opts.InstallMonitor {
		progress.Emit(ctx, "cluster-monitoring", progress.StatusRunning, "installing kube-prometheus-stack via Helm")
		if err := b.installMonitoring(ctx, kubeconfigPath, opts.Provider, domain); err != nil {
			return ports.ClusterState{}, err
		}
		progress.Emit(ctx, "cluster-monitoring", progress.StatusDone, "Prometheus stack installed in namespace monitoring")
	}

	if opts.InstallAgent {
		progress.Emit(ctx, "cluster-agent", progress.StatusRunning, "deploying CAKD Agent manifests")
		if err := b.deployAgent(ctx, kubeconfigPath, opts); err != nil {
			return ports.ClusterState{}, err
		}
		progress.Emit(ctx, "cluster-agent", progress.StatusDone, "CAKD Agent deployed")
	}

	if opts.InstallLogging {
		progress.Emit(ctx, "cluster-logging", progress.StatusRunning, "installing Loki and Promtail via Helm")
		if err := b.installLogging(ctx, kubeconfigPath); err != nil {
			return ports.ClusterState{}, err
		}
		progress.Emit(ctx, "cluster-logging", progress.StatusDone, "Loki and Promtail installed")
	}

	return ports.ClusterState{
		Provider:       opts.Provider,
		Host:           opts.Host,
		Domain:         domain,
		KubeconfigPath: kubeconfigPath,
		SSHKeyPath:     opts.SSHKeyPath,
		SSHUser:        opts.SSHUser,
	}, nil
}

func (b *Bootstrapper) installArgoCD(ctx context.Context, kubeconfig, provider, domain string) error {
	cmds := [][]string{
		{cmdHelm, cmdRepo, cmdAdd, "argo", "https://argoproj.github.io/argo-helm"},
		{cmdHelm, cmdRepo, cmdUpdate, "argo"},
	}
	installArgs := []string{
		cmdHelm, cmdUpgrade, flagInstall, nsArgoCD, "argo/argo-cd",
		flagNamespace, nsArgoCD, flagCreateNamespace,
	}
	if provider == catalog.ClusterAWSEC2 {
		installArgs = append(installArgs,
			"--set", "server.extraArgs[0]=--insecure",
			"--set", "server.ingress.enabled=true",
			"--set", "server.ingress.ingressClassName=traefik",
			"--set", "server.ingress.hosts[0]=argocd."+domain,
			"--set", "dex.enabled=false",
			"--set", "applicationSet.enabled=false",
			"--set", "notifications.enabled=false",
		)
	}
	cmds = append(cmds, installArgs)
	return b.runBatch(ctx, kubeconfig, cmds)
}

func (b *Bootstrapper) installMonitoring(ctx context.Context, kubeconfig, provider, domain string) error {
	amPath, cleanupAM, err := writeEmbeddedTemp("manifests/alertmanager-values.yaml", "alertmanager-values-*.yaml")
	if err != nil {
		return err
	}
	defer cleanupAM()

	gfPath, cleanupGF, err := writeEmbeddedTemp("manifests/grafana-values.yaml", "grafana-values-*.yaml")
	if err != nil {
		return err
	}
	defer cleanupGF()

	cmds := [][]string{
		{cmdHelm, cmdRepo, cmdAdd, "prometheus-community", "https://prometheus-community.github.io/helm-charts"},
		{cmdHelm, cmdRepo, cmdUpdate, "prometheus-community"},
	}
	installArgs := []string{
		cmdHelm, cmdUpgrade, flagInstall, nsMonitoring,
		"prometheus-community/kube-prometheus-stack",
		flagNamespace, nsMonitoring, flagCreateNamespace,
		"-f", amPath,
		"-f", gfPath,
	}
	if provider == catalog.ClusterAWSEC2 {
		installArgs = append(installArgs,
			"--set", "grafana.ingress.enabled=true",
			"--set", "grafana.ingress.ingressClassName=traefik",
			"--set", "grafana.ingress.hosts[0]=grafana."+domain,
		)
	}
	cmds = append(cmds, installArgs)
	return b.runBatch(ctx, kubeconfig, cmds)
}

func (b *Bootstrapper) installLogging(ctx context.Context, kubeconfig string) error {
	lokiPath, cleanupLoki, err := writeEmbeddedTemp("manifests/loki-values.yaml", "loki-values-*.yaml")
	if err != nil {
		return err
	}
	defer cleanupLoki()

	cmds := [][]string{
		{cmdHelm, cmdRepo, cmdAdd, "grafana", "https://grafana.github.io/helm-charts"},
		{cmdHelm, cmdRepo, cmdUpdate, "grafana"},
		{
			cmdHelm, cmdUpgrade, flagInstall, "loki", "grafana/loki",
			flagNamespace, nsMonitoring, flagCreateNamespace,
			"-f", lokiPath,
		},
		{
			cmdHelm, cmdUpgrade, flagInstall, "promtail", "grafana/promtail",
			flagNamespace, nsMonitoring, flagCreateNamespace,
			"--set", "config.lokiAddress=http://loki-gateway.monitoring.svc.cluster.local/loki/api/v1/push",
		},
	}
	return b.runBatch(ctx, kubeconfig, cmds)
}

func (b *Bootstrapper) deployAgent(ctx context.Context, kubeconfig string, opts ports.ClusterBootstrapOptions) error {
	var discordToken string
	geminiKey := opts.GeminiAPIKey
	if b.credStore != nil {
		if creds, err := b.credStore.Get(); err == nil {
			discordToken = creds.DiscordBotToken
			if geminiKey == "" {
				geminiKey = creds.GeminiAPIKey
			}
		}
	}
	webhookURL := opts.DiscordWebhook
	if webhookURL == "" {
		webhookURL = os.Getenv("DISCORD_WEBHOOK_URL")
	}

	tag := "latest"
	if b.agentVersion != "" && b.agentVersion != "dev" && !strings.Contains(b.agentVersion, "dirty") {
		tag = b.agentVersion
	}

	vars := struct {
		DiscordToken   string
		DiscordWebhook string
		GeminiKey      string
		Image          string
	}{
		DiscordToken:   discordToken,
		DiscordWebhook: webhookURL,
		GeminiKey:      geminiKey,
		Image:          "ghcr.io/nguyentin05/cakd-agent:" + tag,
	}

	entries, err := manifestFS.ReadDir("manifests")
	if err != nil {
		return errcode.Wrap(errcode.CodeProvision, "read embedded agent manifests", err)
	}

	var buf bytes.Buffer
	for _, entry := range entries {
		name := entry.Name()
		if name == "alertmanager-values.yaml" || name == "grafana-values.yaml" || name == "loki-values.yaml" {
			continue
		}
		raw, err := manifestFS.ReadFile("manifests/" + name)
		if err != nil {
			return errcode.Wrap(errcode.CodeProvision, "read manifest "+name, err)
		}
		tmpl, err := template.New(name).Parse(string(raw))
		if err != nil {
			return errcode.Wrap(errcode.CodeProvision, "parse manifest "+name, err)
		}
		if err := tmpl.Execute(&buf, vars); err != nil {
			return errcode.Wrap(errcode.CodeProvision, "render manifest "+name, err)
		}
		buf.WriteString("\n---\n")
	}

	tmpFile, err := os.CreateTemp("", "cakd-agent-manifests-*.yaml")
	if err != nil {
		return errcode.Wrap(errcode.CodeProvision, "create agent manifest temp file", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(buf.Bytes()); err != nil {
		_ = tmpFile.Close()
		return errcode.Wrap(errcode.CodeProvision, "write agent manifest temp file", err)
	}
	_ = tmpFile.Close()

	if _, err := b.execCmd(ctx, kubeconfig, "kubectl", "apply", "-f", tmpFile.Name()); err != nil {
		return errcode.Wrap(errcode.CodeProvision, "kubectl apply agent manifests failed", err)
	}
	return nil
}

func (b *Bootstrapper) runBatch(ctx context.Context, kubeconfig string, cmds [][]string) error {
	for _, args := range cmds {
		if _, err := b.execCmd(ctx, kubeconfig, args[0], args[1:]...); err != nil {
			return errcode.Wrap(errcode.CodeProvision, fmt.Sprintf("command %v failed", args), err)
		}
	}
	return nil
}

func (b *Bootstrapper) bootstrapRemoteK3s(ctx context.Context, opts ports.ClusterBootstrapOptions) (string, error) {
	if opts.SSHKeyPath == "" {
		return "", errcode.New(errcode.CodeValidation, "sshKeyPath is required for remote K3s bootstrap")
	}
	pemBytes, err := os.ReadFile(opts.SSHKeyPath)
	if err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, "read SSH private key", err)
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, "parse SSH private key", err)
	}

	user := opts.SSHUser
	if user == "" {
		user = "ubuntu"
	}
	//nolint:gosec
	sshCfg := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	addr := opts.Host + ":22"
	var client *ssh.Client
	for attempt := 1; attempt <= 20; attempt++ {
		client, err = ssh.Dial("tcp", addr, sshCfg)
		if err == nil {
			break
		}
		progress.Emit(ctx, "cluster-ssh", progress.StatusRunning, fmt.Sprintf("waiting for SSH on %s (%d/20)", opts.Host, attempt))
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, "SSH connection failed", err)
	}
	defer client.Close()

	_ = runSSHCommand(client, "[ -f /swapfile ] || (sudo fallocate -l 2G /swapfile && sudo chmod 600 /swapfile && sudo mkswap /swapfile && sudo swapon /swapfile)")

	progress.Emit(ctx, "cluster-k3s", progress.StatusRunning, "installing K3s on "+opts.Host)
	k3sCmd := fmt.Sprintf("curl -sfL https://get.k3s.io | sh -s - --tls-san %s", opts.Host)
	if err := runSSHCommand(client, k3sCmd); err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, "install K3s over SSH", err)
	}

	sess, err := client.NewSession()
	if err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, "create SSH session for kubeconfig", err)
	}
	defer sess.Close()

	waitCmd := "until [ -f /etc/rancher/k3s/k3s.yaml ]; do sleep 2; done && sudo cat /etc/rancher/k3s/k3s.yaml"
	kubeBytes, err := sess.CombinedOutput(waitCmd)
	if err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, "fetch remote k3s.yaml", err)
	}

	kubeStr := strings.Replace(string(kubeBytes), "127.0.0.1", opts.Host, 1)
	kubeStr = strings.Replace(kubeStr, "localhost", opts.Host, 1)

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	kubeconfigPath := filepath.Join(home, ".kube", "cakd-config")
	if err := os.MkdirAll(filepath.Dir(kubeconfigPath), 0o750); err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, "create .kube directory", err)
	}
	if err := os.WriteFile(kubeconfigPath, []byte(kubeStr), 0o600); err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, "write local kubeconfig", err)
	}
	return kubeconfigPath, nil
}

func runSSHCommand(client *ssh.Client, cmd string) error {
	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()
	return sess.Run(cmd)
}

func writeEmbeddedTemp(embedPath, pattern string) (string, func(), error) {
	raw, err := manifestFS.ReadFile(embedPath)
	if err != nil {
		return "", nil, errcode.Wrap(errcode.CodeProvision, "read embedded manifest "+embedPath, err)
	}
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", nil, errcode.Wrap(errcode.CodeProvision, "create temp values file", err)
	}
	if _, err := f.Write(raw); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", nil, errcode.Wrap(errcode.CodeProvision, "write temp values file", err)
	}
	_ = f.Close()
	return f.Name(), func() { _ = os.Remove(f.Name()) }, nil
}
