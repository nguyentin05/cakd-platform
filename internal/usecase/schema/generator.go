package schema

import (
	"encoding/json"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
)

// GenerateJSONSchema returns the formatted JSON Schema (Draft 2020-12) for platform.yaml.
func GenerateJSONSchema() ([]byte, error) {
	doc := map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"$id":                  "https://cakd.io/schemas/platform.schema.json",
		"title":                "CAKD Platform Project Specification",
		"description":          "Declarative schema for CAKD Platform (cakd.io/v1alpha1)",
		"type":                 "object",
		"required":             []string{"apiVersion", "kind", "metadata", "spec"},
		"additionalProperties": false,
		"properties": map[string]any{
			"apiVersion": map[string]any{
				"type":    "string",
				"enum":    catalog.APIVersions,
				"default": "cakd.io/v1alpha1",
			},
			"kind": map[string]any{
				"type":    "string",
				"enum":    catalog.Kinds,
				"default": "Project",
			},
			"metadata": map[string]any{
				"type":                 "object",
				"required":             []string{"name", "owner"},
				"additionalProperties": false,
				"properties": map[string]any{
					"name":        map[string]any{"type": "string", "minLength": 1},
					"owner":       map[string]any{"type": "string", "minLength": 1},
					"environment": map[string]any{"type": "string", "default": catalog.Defaults.Environment},
				},
			},
			"spec": map[string]any{
				"type":                 "object",
				"required":             []string{"services"},
				"additionalProperties": false,
				"properties": map[string]any{
					"cluster": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"provider":     map[string]any{"type": "string", "enum": catalog.ClusterProviders, "default": catalog.Defaults.ClusterProvider},
							"region":       map[string]any{"type": "string", "default": catalog.Defaults.Region},
							"instanceType": map[string]any{"type": "string", "enum": catalog.InstanceTypes, "default": catalog.Defaults.InstanceType},
							"diskSize":     map[string]any{"type": "integer", "default": catalog.Defaults.DiskSize},
							"domain":       map[string]any{"type": "string"},
							"host":         map[string]any{"type": "string"},
							"sshKeyPath":   map[string]any{"type": "string"},
							"sshUser":      map[string]any{"type": "string"},
							"addons": map[string]any{
								"type":                 "object",
								"additionalProperties": false,
								"properties": map[string]any{
									"argocd":     map[string]any{"type": "boolean"},
									"monitoring": map[string]any{"type": "boolean"},
									"logging":    map[string]any{"type": "boolean"},
									"agent":      map[string]any{"type": "boolean"},
								},
							},
						},
					},
					"providers": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"versionControl": map[string]any{"type": "string", "enum": catalog.VersionControlTypes, "default": catalog.Defaults.VersionControl},
							"ci":             map[string]any{"type": "string", "enum": catalog.CITypes, "default": catalog.Defaults.CI},
							"cd":             map[string]any{"type": "string", "enum": catalog.CDTypes, "default": catalog.Defaults.CD},
							"notification":   map[string]any{"type": "string", "enum": catalog.NotificationTypes},
							"llm":            map[string]any{"type": "string", "enum": catalog.LLMTypes},
							"monitoring":     map[string]any{"type": "string", "enum": catalog.MonitoringTypes},
							"logging":        map[string]any{"type": "string", "enum": catalog.LoggingTypes},
						},
					},
					"observability": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"alerting": map[string]any{"type": "boolean"},
							"ai": map[string]any{
								"type":                 "object",
								"required":             []string{"model"},
								"additionalProperties": false,
								"properties": map[string]any{
									"model": map[string]any{"type": "string"},
								},
							},
						},
					},
					"backing": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type":                 "object",
							"required":             []string{"name", "type"},
							"additionalProperties": false,
							"properties": map[string]any{
								"name":    map[string]any{"type": "string"},
								"type":    map[string]any{"type": "string", "enum": catalog.BackingTypes},
								"version": map[string]any{"type": "string"},
								"storage": map[string]any{"type": "string", "default": catalog.Defaults.Storage},
							},
						},
					},
					"services": map[string]any{
						"type":     "array",
						"minItems": 1,
						"items": map[string]any{
							"type":                 "object",
							"required":             []string{"name", "language"},
							"additionalProperties": false,
							"properties": map[string]any{
								"name":              map[string]any{"type": "string"},
								"language":          map[string]any{"type": "string", "enum": catalog.ServiceLanguages},
								"version":           map[string]any{"type": "string", "enum": catalog.ValidJavaVersions},
								"springBootVersion": map[string]any{"type": "string", "enum": catalog.ValidSpringBootVersions},
								"projectBuild":      map[string]any{"type": "string", "enum": catalog.ValidProjectBuilds},
								"groupId":           map[string]any{"type": "string"},
								"artifactId":        map[string]any{"type": "string"},
								"description":       map[string]any{"type": "string"},
								"packageName":       map[string]any{"type": "string"},
								"packaging":         map[string]any{"type": "string", "enum": catalog.ValidPackaging},
								"configuration":     map[string]any{"type": "string", "enum": catalog.SpringConfigFormats},
								"dependencies": map[string]any{
									"type":  "array",
									"items": map[string]any{"type": "string", "enum": catalog.ValidSpringDependencies},
								},
								"replicas": map[string]any{"type": "integer", "default": catalog.Defaults.Replicas},
								"resources": map[string]any{
									"type":                 "object",
									"additionalProperties": false,
									"properties": map[string]any{
										"cpu":    map[string]any{"type": "string", "default": catalog.Defaults.CPU},
										"memory": map[string]any{"type": "string", "default": catalog.Defaults.Memory},
									},
								},
								"uses": map[string]any{
									"type":  "array",
									"items": map[string]any{"type": "string"},
								},
							},
						},
					},
				},
			},
		},
	}
	return json.MarshalIndent(doc, "", "  ")
}
