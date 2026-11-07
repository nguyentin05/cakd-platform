package configfile

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
)

type rawEnvelope struct {
	APIVersion    string                     `yaml:"apiVersion"`
	Kind          string                     `yaml:"kind"`
	Metadata      project.Metadata           `yaml:"metadata"`
	Spec          *project.Spec              `yaml:"spec,omitempty"`
	Infra         *project.ClusterSpec       `yaml:"infra,omitempty"`
	Cluster       *project.ClusterSpec       `yaml:"cluster,omitempty"`
	Providers     *project.ProvidersSpec     `yaml:"providers,omitempty"`
	Observability *project.ObservabilitySpec `yaml:"observability,omitempty"`
	Backing       []project.BackingSpec      `yaml:"backing,omitempty"`
	Services      []project.ServiceSpec      `yaml:"services,omitempty"`
}

// Loader implements ports.SpecRepository for reading and writing platform.yaml files.
type Loader struct{}

// NewLoader creates a new YAML configuration Loader.
func NewLoader() *Loader {
	return &Loader{}
}

// LoadFile reads a YAML file from disk and normalizes it into a domain Project.
func (l *Loader) LoadFile(path string) (*project.Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errcode.Wrap(errcode.CodeNotFound, "cannot read configuration file", err)
	}
	return l.LoadBytes(data)
}

// LoadBytes parses raw YAML bytes (either canonical cakd.io/v1alpha1 or legacy platform.dev/v1alpha1)
// into a normalized domain Project.
func (l *Loader) LoadBytes(data []byte) (*project.Project, error) {
	var raw rawEnvelope
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, errcode.Wrap(errcode.CodeValidation, "invalid YAML syntax", err)
	}

	proj := &project.Project{
		APIVersion: raw.APIVersion,
		Kind:       raw.Kind,
		Metadata:   raw.Metadata,
	}
	if raw.Spec != nil {
		proj.Spec = *raw.Spec
	}

	// Normalize legacy top-level blocks into Spec when Spec fields are unset.
	if proj.Spec.Cluster == nil {
		if raw.Cluster != nil {
			proj.Spec.Cluster = raw.Cluster
		} else if raw.Infra != nil {
			proj.Spec.Cluster = raw.Infra
		}
	}
	if proj.Spec.Providers == nil && raw.Providers != nil {
		proj.Spec.Providers = raw.Providers
	}
	if proj.Spec.Observability == nil && raw.Observability != nil {
		proj.Spec.Observability = raw.Observability
	}
	if len(proj.Spec.Backing) == 0 && len(raw.Backing) > 0 {
		proj.Spec.Backing = raw.Backing
	}
	if len(proj.Spec.Services) == 0 && len(raw.Services) > 0 {
		proj.Spec.Services = raw.Services
	}

	return proj, nil
}

// MarshalCanonical serializes a Project using the canonical cakd.io/v1alpha1 schema and spec block.
func (l *Loader) MarshalCanonical(proj *project.Project) ([]byte, error) {
	out := *proj
	out.APIVersion = project.CanonicalAPIVersion
	if out.Kind == "" {
		out.Kind = project.KindProject
	}
	data, err := yaml.Marshal(&out)
	if err != nil {
		return nil, errcode.Wrap(errcode.CodeValidation, "failed to marshal project YAML", err)
	}
	return data, nil
}

// SaveFile writes a Project in canonical YAML format to path.
func (l *Loader) SaveFile(path string, proj *project.Project) error {
	data, err := l.MarshalCanonical(proj)
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return errcode.Wrap(errcode.CodeExternal, "failed to create parent directory", err)
		}
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return errcode.Wrap(errcode.CodeExternal, "failed to write configuration file", err)
	}
	return nil
}
