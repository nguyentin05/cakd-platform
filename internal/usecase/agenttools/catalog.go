package agenttools

import (
	"context"
	"encoding/json"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/usecase/createproject"
	"github.com/nguyentin05/cakd-platform/internal/usecase/initcluster"
	"github.com/nguyentin05/cakd-platform/internal/usecase/observe"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
	"github.com/nguyentin05/cakd-platform/internal/usecase/schema"
	"github.com/nguyentin05/cakd-platform/internal/usecase/validate"
)

// ToolAnnotations describes MCP safety and idempotency hints for AI agents.
type ToolAnnotations struct {
	ReadOnlyHint    bool `json:"readOnlyHint"`
	DestructiveHint bool `json:"destructiveHint"`
	IdempotentHint  bool `json:"idempotentHint"`
}

// ToolDefinition defines a single MCP tool exposed by CAKD Platform.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema map[string]any  `json:"inputSchema"`
	Annotations ToolAnnotations `json:"annotations"`
}

// ResourceDefinition defines a single MCP resource exposed by CAKD Platform.
type ResourceDefinition struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// Catalog dispatches MCP tool and resource calls to CAKD application use cases.
type Catalog struct {
	specRepo  ports.SpecRepository
	validator *validate.Service
	createUC  *createproject.Service
	initUC    *initcluster.Service
	observeUC *observe.Service
}

// NewCatalog creates an MCP agent tool Catalog wired to core use cases.
func NewCatalog(
	specRepo ports.SpecRepository,
	validator *validate.Service,
	createUC *createproject.Service,
	initUC *initcluster.Service,
	observeUC *observe.Service,
) *Catalog {
	return &Catalog{
		specRepo:  specRepo,
		validator: validator,
		createUC:  createUC,
		initUC:    initUC,
		observeUC: observeUC,
	}
}

// ListResources returns all MCP resources served by CAKD Platform.
func (c *Catalog) ListResources() []ResourceDefinition {
	return []ResourceDefinition{
		{
			URI:         "cakd://schema/project",
			Name:        "CAKD Project JSON Schema",
			Description: "JSON Schema (Draft 2020-12) for platform.yaml (cakd.io/v1alpha1)",
			MimeType:    "application/json",
		},
		{
			URI:         "cakd://catalog/providers",
			Name:        "Supported Providers & Defaults",
			Description: "Catalog of supported cluster providers, backing services, languages, and defaults",
			MimeType:    "application/json",
		},
	}
}

// ReadResource reads the content and MIME type of an MCP resource by URI.
func (c *Catalog) ReadResource(_ context.Context, uri string) (content string, mimeType string, err error) {
	switch uri {
	case "cakd://schema/project":
		data, err := schema.GenerateJSONSchema()
		if err != nil {
			return "", "", err
		}
		return string(data), "application/json", nil
	case "cakd://catalog/providers":
		out, err := c.catalogJSON()
		if err != nil {
			return "", "", err
		}
		return out, "application/json", nil
	default:
		return "", "", errcode.New(errcode.CodeNotFound, "unknown MCP resource URI: "+uri)
	}
}

// ListTools returns all MCP tools exposed to AI agents.
func (c *Catalog) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "cakd_list_catalog",
			Description: "List supported cluster providers, backing services, Java/Spring Boot versions, and platform defaults.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
			Annotations: ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
		},
		{
			Name:        "cakd_validate_spec",
			Description: "Validate a platform.yaml specification (either raw YAML content or file path) and return normalized defaults.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"yamlContent": map[string]any{"type": "string", "description": "Raw YAML content of platform.yaml"},
					"path":        map[string]any{"type": "string", "description": "Path to platform.yaml on disk"},
				},
			},
			Annotations: ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
		},
		{
			Name:        "cakd_generate_spec",
			Description: "Generate a canonical cakd.io/v1alpha1 platform.yaml from high-level project parameters.",
			InputSchema: map[string]any{
				"type":     "object",
				"required": []string{"name", "owner", "serviceName"},
				"properties": map[string]any{
					"name":        map[string]any{"type": "string", "description": "Project metadata.name"},
					"owner":       map[string]any{"type": "string", "description": "GitHub owner/organization"},
					"serviceName": map[string]any{"type": "string", "description": "Primary microservice name"},
					"provider":    map[string]any{"type": "string", "description": "Cluster provider: local or aws-ec2"},
					"backingName": map[string]any{"type": "string", "description": "Optional PostgreSQL backing resource name"},
					"outputPath":  map[string]any{"type": "string", "description": "Optional file path to save platform.yaml"},
				},
			},
			Annotations: ToolAnnotations{ReadOnlyHint: false, IdempotentHint: true},
		},
		{
			Name:        "cakd_scaffold_only",
			Description: "Render microservice skeletons, Dockerfiles, CI workflows, and Helm charts locally without provisioning cloud resources.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":      map[string]any{"type": "string", "description": "Path to platform.yaml"},
					"outputDir": map[string]any{"type": "string", "description": "Target output directory"},
					"force":     map[string]any{"type": "boolean", "description": "Overwrite output directory if it exists"},
				},
			},
			Annotations: ToolAnnotations{ReadOnlyHint: false, DestructiveHint: false, IdempotentHint: true},
		},
		{
			Name:        "cakd_create_project",
			Description: "Execute full project bootstrap: validate, scaffold, provision GitHub repo via Terraform, push git, and register ArgoCD.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":  map[string]any{"type": "string", "description": "Path to platform.yaml"},
					"force": map[string]any{"type": "boolean", "description": "Overwrite output directory if it exists"},
				},
			},
			Annotations: ToolAnnotations{ReadOnlyHint: false, DestructiveHint: true},
		},
		{
			Name:        "cakd_init_cluster",
			Description: "Initialize target Kubernetes cluster (local or AWS EC2 K3s) and install ArgoCD, Prometheus, Loki, and CAKD Agent.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":       map[string]any{"type": "string", "description": "Optional path to platform.yaml"},
					"argocd":     map[string]any{"type": "boolean"},
					"monitoring": map[string]any{"type": "boolean"},
					"logging":    map[string]any{"type": "boolean"},
				},
			},
			Annotations: ToolAnnotations{ReadOnlyHint: false, DestructiveHint: true},
		},
		{
			Name:        "cakd_observe_workload",
			Description: "Collect Prometheus metrics and Loki logs for a namespace and generate an AI root-cause diagnosis.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"namespace": map[string]any{"type": "string", "description": "Target Kubernetes namespace"},
					"path":      map[string]any{"type": "string", "description": "Optional path to platform.yaml"},
				},
			},
			Annotations: ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
		},
	}
}

// CallTool executes an MCP tool by name with JSON arguments and returns a formatted text/JSON response.
//
//nolint:gocyclo
func (c *Catalog) CallTool(ctx context.Context, name string, args map[string]any) (string, error) {
	switch name {
	case "cakd_list_catalog":
		return c.catalogJSON()

	case "cakd_validate_spec":
		yamlContent, _ := args["yamlContent"].(string)
		path, _ := args["path"].(string)
		var proj *project.Project
		var err error
		if yamlContent != "" {
			proj, err = c.specRepo.LoadBytes([]byte(yamlContent))
		} else {
			if path == "" {
				path = "platform.yaml"
			}
			proj, err = c.specRepo.LoadFile(path)
		}
		if err != nil {
			return "", err
		}
		if err := c.validator.ValidateAndDefault(proj); err != nil {
			return "", err
		}
		resp := map[string]any{
			"valid":      true,
			"normalized": proj,
		}
		b, _ := json.MarshalIndent(resp, "", "  ")
		return string(b), nil

	case "cakd_generate_spec":
		projName, _ := args["name"].(string)
		owner, _ := args["owner"].(string)
		svcName, _ := args["serviceName"].(string)
		provider, _ := args["provider"].(string)
		backingName, _ := args["backingName"].(string)
		outPath, _ := args["outputPath"].(string)

		if provider == "" {
			provider = catalog.ClusterLocal
		}
		proj := &project.Project{
			APIVersion: project.CanonicalAPIVersion,
			Kind:       project.KindProject,
			Metadata: project.Metadata{
				Name:  projName,
				Owner: owner,
			},
			Spec: project.Spec{
				Cluster: &project.ClusterSpec{Provider: provider},
				Services: []project.ServiceSpec{
					{Name: svcName, Language: catalog.LangJavaSpringBoot},
				},
			},
		}
		if backingName != "" {
			proj.Spec.Backing = []project.BackingSpec{
				{Name: backingName, Type: catalog.BackingPostgreSQL},
			}
			proj.Spec.Services[0].Uses = []string{backingName}
		}
		if err := c.validator.ValidateAndDefault(proj); err != nil {
			return "", err
		}
		if outPath != "" {
			if err := c.specRepo.SaveFile(outPath, proj); err != nil {
				return "", err
			}
		}
		b, err := c.specRepo.MarshalCanonical(proj)
		if err != nil {
			return "", err
		}
		return string(b), nil

	case "cakd_scaffold_only":
		path, _ := args["path"].(string)
		if path == "" {
			path = "platform.yaml"
		}
		outDir, _ := args["outputDir"].(string)
		force, _ := args["force"].(bool)
		res, err := c.createUC.Execute(ctx, createproject.Request{
			SpecPath:     path,
			OutputDir:    outDir,
			Force:        force,
			ScaffoldOnly: true,
		})
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), nil

	case "cakd_create_project":
		path, _ := args["path"].(string)
		if path == "" {
			path = "platform.yaml"
		}
		force, _ := args["force"].(bool)
		res, err := c.createUC.Execute(ctx, createproject.Request{
			SpecPath: path,
			Force:    force,
		})
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), nil

	case "cakd_init_cluster":
		path, _ := args["path"].(string)
		argocd, _ := args["argocd"].(bool)
		monitoring, _ := args["monitoring"].(bool)
		logging, _ := args["logging"].(bool)
		st, err := c.initUC.Execute(ctx, initcluster.Request{
			SpecPath:   path,
			ArgoCD:     argocd,
			Monitoring: monitoring,
			Logging:    logging,
		})
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(st, "", "  ")
		return string(b), nil

	case "cakd_observe_workload":
		ns, _ := args["namespace"].(string)
		path, _ := args["path"].(string)
		rep, err := c.observeUC.Execute(ctx, observe.Request{
			SpecPath:  path,
			Namespace: ns,
		})
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(rep, "", "  ")
		return string(b), nil

	default:
		return "", errcode.New(errcode.CodeNotFound, "unknown MCP tool: "+name)
	}
}

func (c *Catalog) catalogJSON() (string, error) {
	payload := map[string]any{
		"apiVersions":          catalog.APIVersions,
		"clusterProviders":     catalog.ClusterProviders,
		"instanceTypes":        catalog.InstanceTypes,
		"versionControl":       catalog.VersionControlTypes,
		"ci":                   catalog.CITypes,
		"cd":                   catalog.CDTypes,
		"notification":         catalog.NotificationTypes,
		"llm":                  catalog.LLMTypes,
		"monitoring":           catalog.MonitoringTypes,
		"logging":              catalog.LoggingTypes,
		"languages":            catalog.ServiceLanguages,
		"backingTypes":         catalog.BackingTypes,
		"javaVersions":         catalog.ValidJavaVersions,
		"springBootVersions":   catalog.ValidSpringBootVersions,
		"projectBuilds":        catalog.ValidProjectBuilds,
		"packaging":            catalog.ValidPackaging,
		"springConfigFormats":  catalog.SpringConfigFormats,
		"defaults":             catalog.Defaults,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
