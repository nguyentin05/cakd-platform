package catalog

// CatalogIndex represents the top-level catalog registry indexing available providers.
type CatalogIndex struct {
	APIVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Providers  []ProviderDescriptor `json:"providers" yaml:"providers"`
}

// ProviderDescriptor provides lightweight metadata for provider discovery.
type ProviderDescriptor struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Category    string `json:"category" yaml:"category"`
	Description string `json:"description" yaml:"description"`
	Version     string `json:"version,omitempty" yaml:"version,omitempty"`
}

// ProviderManifest describes the complete declarative configuration and lifecycle of a provider.
type ProviderManifest struct {
	ID            string             `json:"id" yaml:"id"`
	Name          string             `json:"name" yaml:"name"`
	Category      string             `json:"category" yaml:"category"`
	Description   string             `json:"description" yaml:"description"`
	Version       string             `json:"version,omitempty" yaml:"version,omitempty"`
	Introspection *IntrospectionSpec `json:"introspection,omitempty" yaml:"introspection,omitempty"`
	Steps         []StepSpec         `json:"steps,omitempty" yaml:"steps,omitempty"`
}

// IntrospectionSpec defines HTTP endpoints to query dynamic configuration options from the provider.
type IntrospectionSpec struct {
	Fetch FetchSpec `json:"fetch" yaml:"fetch"`
}

// FetchSpec specifies the HTTP request for dynamic introspection.
type FetchSpec struct {
	URL     string            `json:"url" yaml:"url"`
	Method  string            `json:"method,omitempty" yaml:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// StepSpec defines an interactive configuration step or option prompt.
type StepSpec struct {
	ID      string         `json:"id" yaml:"id"`
	Title   string         `json:"title" yaml:"title"`
	Type    string         `json:"type,omitempty" yaml:"type,omitempty"`
	Default string         `json:"default,omitempty" yaml:"default,omitempty"`
	Extract string         `json:"extract,omitempty" yaml:"extract,omitempty"`
	Options []ChoiceOption `json:"options,omitempty" yaml:"options,omitempty"`
}

// ChoiceOption represents a selectable option in an interactive step.
type ChoiceOption struct {
	Label string `json:"label" yaml:"label"`
	Value string `json:"value" yaml:"value"`
}
