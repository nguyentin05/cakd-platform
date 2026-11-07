package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/composition"
	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
	"github.com/nguyentin05/cakd-platform/internal/usecase/createproject"
)

// Options configures the interactive TUI project creation wizard.
type Options struct {
	OutputPath string
	Force      bool
}

// RunWizard runs an interactive step-by-step terminal wizard that builds a canonical
// platform.yaml without requiring manual YAML authoring, and optionally executes project bootstrap.
func RunWizard(ctx context.Context, in io.Reader, out io.Writer, c *composition.Container, opts Options) error {
	scanner := bufio.NewScanner(in)
	outPath := opts.OutputPath
	if outPath == "" {
		outPath = "platform.yaml"
	}

	_, _ = fmt.Fprintln(out, "╭────────────────────────────────────────────────────────╮")
	_, _ = fmt.Fprintln(out, "│      CAKD Platform — Interactive Project Wizard        │")
	_, _ = fmt.Fprintln(out, "╰────────────────────────────────────────────────────────╯")

	projName := askDefault(scanner, out, "1. Project name", "my-platform")
	owner := askDefault(scanner, out, "2. GitHub owner / organization", "my-org")
	provider := askChoice(scanner, out, "3. Cluster provider", catalog.ClusterProviders, catalog.Defaults.ClusterProvider)

	clusterSpec := &project.ClusterSpec{
		Provider: provider,
	}
	if provider == catalog.ClusterAWSEC2 {
		clusterSpec.Region = askDefault(scanner, out, "   AWS Region", catalog.Defaults.Region)
		clusterSpec.InstanceType = askChoice(scanner, out, "   EC2 Instance Type", catalog.InstanceTypes, catalog.Defaults.InstanceType)
		clusterSpec.DiskSize = catalog.Defaults.DiskSize
	}

	svcName := askDefault(scanner, out, "4. Primary microservice name", "api-service")
	javaVer := askChoice(scanner, out, "5. Java version", catalog.ValidJavaVersions, catalog.Defaults.LanguageVersion[catalog.LangJavaSpringBoot])
	buildTool := askChoice(scanner, out, "6. Build tool", []string{"maven-project", "gradle-project"}, "maven-project")

	var backing []project.BackingSpec
	var uses []string
	if askYesNo(scanner, out, "7. Provision PostgreSQL backing database? [y/N]", false) {
		dbName := askDefault(scanner, out, "   PostgreSQL resource name", svcName+"-db")
		backing = append(backing, project.BackingSpec{
			Name:    dbName,
			Type:    catalog.BackingPostgreSQL,
			Version: catalog.Defaults.DBVersion[catalog.BackingPostgreSQL],
			Storage: catalog.Defaults.Storage,
		})
		uses = append(uses, dbName)
	}

	proj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata: project.Metadata{
			Name:        projName,
			Owner:       owner,
			Environment: catalog.Defaults.Environment,
		},
		Spec: project.Spec{
			Cluster: clusterSpec,
			Providers: &project.ProvidersSpec{
				VersionControl: catalog.Defaults.VersionControl,
				CI:             catalog.Defaults.CI,
				CD:             catalog.Defaults.CD,
			},
			Backing: backing,
			Services: []project.ServiceSpec{
				{
					Name:         svcName,
					Language:     catalog.LangJavaSpringBoot,
					Version:      javaVer,
					ProjectBuild: buildTool,
					Uses:         uses,
				},
			},
		},
	}

	if err := c.ValidateUC.ValidateAndDefault(proj); err != nil {
		return err
	}

	yamlBytes, err := c.SpecRepo.MarshalCanonical(proj)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(out, "\n── Generated platform.yaml Preview ──────────────────────")
	_, _ = fmt.Fprint(out, string(yamlBytes))
	_, _ = fmt.Fprintln(out, "─────────────────────────────────────────────────────────")

	if err := c.SpecRepo.SaveFile(outPath, proj); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "✔ Saved configuration to %s\n", outPath)

	if askYesNo(scanner, out, "8. Run 'cakd create' now? [y/N]", false) {
		ctxWithSink := progress.WithSink(ctx, progress.SinkFunc(func(ev progress.Event) {
			_, _ = fmt.Fprintf(out, "  [%s] (%s) %s\n", ev.Phase, ev.Status, ev.Message)
		}))
		res, err := c.CreateUC.Execute(ctxWithSink, createproject.Request{
			Project: proj,
			Force:   opts.Force,
		})
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "✔ Project created at %s (domain: %s)\n", res.OutputDir, res.Domain)
	}

	return nil
}

func askDefault(scanner *bufio.Scanner, out io.Writer, label, defaultVal string) string {
	_, _ = fmt.Fprintf(out, "%s [%s]: ", label, defaultVal)
	if scanner.Scan() {
		if val := strings.TrimSpace(scanner.Text()); val != "" {
			return val
		}
	}
	return defaultVal
}

func askChoice(scanner *bufio.Scanner, out io.Writer, label string, options []string, defaultVal string) string {
	_, _ = fmt.Fprintf(out, "%s (%s) [%s]: ", label, strings.Join(options, "/"), defaultVal)
	if scanner.Scan() {
		if val := strings.TrimSpace(scanner.Text()); val != "" {
			return val
		}
	}
	return defaultVal
}

func askYesNo(scanner *bufio.Scanner, out io.Writer, label string, defaultYes bool) bool {
	_, _ = fmt.Fprintf(out, "%s: ", label)
	if scanner.Scan() {
		val := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if val == "y" || val == "yes" {
			return true
		}
		if val == "n" || val == "no" {
			return false
		}
	}
	return defaultYes
}
