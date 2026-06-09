package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func deployAgent() error {
	fmt.Println("Step: Building CAKD Agent Docker Image...")
	buildCmd := exec.Command("docker", "build", "-t", "cakd-agent:latest", ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build docker image: %w", err)
	}

	fmt.Println("Step: Loading Image into Minikube...")
	loadCmd := exec.Command("minikube", "image", "load", "cakd-agent:latest")
	loadCmd.Stdout = os.Stdout
	loadCmd.Stderr = os.Stderr
	if err := loadCmd.Run(); err != nil {
		return fmt.Errorf("failed to load image into minikube: %w", err)
	}

	discordToken := os.Getenv("DISCORD_BOT_TOKEN")
	discordWebhook := os.Getenv("DISCORD_WEBHOOK_URL")
	geminiKey := os.Getenv("GEMINI_API_KEY")

	if discordToken == "" || geminiKey == "" {
		fmt.Println("Warning: DISCORD_BOT_TOKEN or GEMINI_API_KEY is not set. Agent might not work fully.")
	}

	manifest := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: cakd-agent-secret
  namespace: monitoring
type: Opaque
stringData:
  DISCORD_BOT_TOKEN: "%s"
  DISCORD_WEBHOOK_URL: "%s"
  GEMINI_API_KEY: "%s"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cakd-webhooks
  namespace: monitoring
data:
  webhooks.json: |
    {}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cakd-agent
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cakd-agent
  template:
    metadata:
      labels:
        app: cakd-agent
    spec:
      containers:
      - name: agent
        image: cakd-agent:latest
        imagePullPolicy: Never
        ports:
        - containerPort: 8080
        envFrom:
        - secretRef:
            name: cakd-agent-secret
        volumeMounts:
        - name: webhooks-volume
          mountPath: /etc/cakd
      volumes:
      - name: webhooks-volume
        configMap:
          name: cakd-webhooks
---
apiVersion: v1
kind: Service
metadata:
  name: cakd-agent
  namespace: monitoring
spec:
  selector:
    app: cakd-agent
  ports:
  - port: 8080
    targetPort: 8080
`, discordToken, discordWebhook, geminiKey)

	fmt.Println("Step: Deploying CAKD Agent to Kubernetes...")
	kubectlCmd := exec.Command("kubectl", "apply", "-f", "-")
	kubectlCmd.Stdin = strings.NewReader(manifest)
	kubectlCmd.Stdout = os.Stdout
	kubectlCmd.Stderr = os.Stderr

	if err := kubectlCmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy agent manifests: %w", err)
	}

	return nil
}
