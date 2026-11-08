package schema

import (
	"encoding/json"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
)

const (
	keyType                 = "type"
	keyRequired             = "required"
	keyAdditionalProperties = "additionalProperties"
	keyProperties           = "properties"
	keyEnum                 = "enum"
	keyDefault              = "default"
	keyItems                = "items"
	keyName                 = "name"
	keyDescription          = "description"
	typeObject              = "object"
	typeString              = "string"
	typeBoolean             = "boolean"
	typeArray               = "array"
	typeInteger             = "integer"
)

// GenerateJSONSchema returns the formatted JSON Schema (Draft 2020-12) for platform.yaml.
//
//nolint:goconst
func GenerateJSONSchema() ([]byte, error) {
	doc := map[string]any{
		"$schema":               "https://json-schema.org/draft/2020-12/schema",
		"$id":                   "https://cakd.io/schemas/platform.schema.json",
		"title":                 "CAKD Platform Project Specification",
		"description":           "Declarative schema for CAKD Platform (cakd.io/v1alpha1)",
		keyType:                 typeObject,
		keyRequired:             []string{"apiVersion", "kind", "metadata", "spec"},
		keyAdditionalProperties: false,
		keyProperties: map[string]any{
			"apiVersion": map[string]any{
				keyType:    typeString,
				keyEnum:    catalog.APIVersions,
				keyDefault: "cakd.io/v1alpha1",
			},
			"kind": map[string]any{
				keyType:    typeString,
				keyEnum:    catalog.Kinds,
				keyDefault: "Project",
			},
			"metadata": map[string]any{
				keyType:                 typeObject,
				keyRequired:             []string{keyName, "owner"},
				keyAdditionalProperties: false,
				keyProperties: map[string]any{
					keyName:       map[string]any{keyType: typeString, "minLength": 1},
					"owner":       map[string]any{keyType: typeString, "minLength": 1},
					"environment": map[string]any{keyType: typeString, keyDefault: catalog.Defaults.Environment},
				},
			},
			"spec": map[string]any{
				keyType:                 typeObject,
				keyRequired:             []string{"services"},
				keyAdditionalProperties: false,
				keyProperties: map[string]any{
					"cluster": map[string]any{
						keyType:                 typeObject,
						keyAdditionalProperties: false,
						keyProperties: map[string]any{
							"provider":     map[string]any{keyType: typeString, keyEnum: catalog.ClusterProviders, keyDefault: catalog.Defaults.ClusterProvider},
							"region":       map[string]any{keyType: typeString, keyDefault: catalog.Defaults.Region},
							"instanceType": map[string]any{keyType: typeString, keyEnum: catalog.InstanceTypes, keyDefault: catalog.Defaults.InstanceType},
							"diskSize":     map[string]any{keyType: typeInteger, keyDefault: catalog.Defaults.DiskSize},
							"domain":       map[string]any{keyType: typeString},
							"host":         map[string]any{keyType: typeString},
							"sshKeyPath":   map[string]any{keyType: typeString},
							"sshUser":      map[string]any{keyType: typeString},
							"addons": map[string]any{
								keyType:                 typeObject,
								keyAdditionalProperties: false,
								keyProperties: map[string]any{
									"argocd":     map[string]any{keyType: typeBoolean},
									"monitoring": map[string]any{keyType: typeBoolean},
									"logging":    map[string]any{keyType: typeBoolean},
									"agent":      map[string]any{keyType: typeBoolean},
								},
							},
						},
					},
					"providers": map[string]any{
						keyType:                 typeObject,
						keyAdditionalProperties: false,
						keyProperties: map[string]any{
							"versionControl": map[string]any{keyType: typeString, keyEnum: catalog.VersionControlTypes, keyDefault: catalog.Defaults.VersionControl},
							"ci":             map[string]any{keyType: typeString, keyEnum: catalog.CITypes, keyDefault: catalog.Defaults.CI},
							"cd":             map[string]any{keyType: typeString, keyEnum: catalog.CDTypes, keyDefault: catalog.Defaults.CD},
							"notification":   map[string]any{keyType: typeString, keyEnum: catalog.NotificationTypes},
							"llm":            map[string]any{keyType: typeString, keyEnum: catalog.LLMTypes},
							"monitoring":     map[string]any{keyType: typeString, keyEnum: catalog.MonitoringTypes},
							"logging":        map[string]any{keyType: typeString, keyEnum: catalog.LoggingTypes},
						},
					},
					"observability": map[string]any{
						keyType:                 typeObject,
						keyAdditionalProperties: false,
						keyProperties: map[string]any{
							"alerting": map[string]any{keyType: typeBoolean},
							"ai": map[string]any{
								keyType:                 typeObject,
								keyRequired:             []string{"model"},
								keyAdditionalProperties: false,
								keyProperties: map[string]any{
									"model": map[string]any{keyType: typeString},
								},
							},
						},
					},
					"backing": map[string]any{
						keyType: typeArray,
						keyItems: map[string]any{
							keyType:                 typeObject,
							keyRequired:             []string{keyName, keyType},
							keyAdditionalProperties: false,
							keyProperties: map[string]any{
								keyName:   map[string]any{keyType: typeString},
								keyType:   map[string]any{keyType: typeString, keyEnum: catalog.BackingTypes},
								"version": map[string]any{keyType: typeString},
								"storage": map[string]any{keyType: typeString, keyDefault: catalog.Defaults.Storage},
							},
						},
					},
					"services": map[string]any{
						keyType:    typeArray,
						"minItems": 1,
						keyItems: map[string]any{
							keyType:                 typeObject,
							keyRequired:             []string{keyName, "language"},
							keyAdditionalProperties: false,
							keyProperties: map[string]any{
								keyName:             map[string]any{keyType: typeString},
								"language":          map[string]any{keyType: typeString, keyEnum: catalog.ServiceLanguages},
								"version":           map[string]any{keyType: typeString, keyEnum: catalog.ValidJavaVersions},
								"springBootVersion": map[string]any{keyType: typeString, keyEnum: catalog.ValidSpringBootVersions},
								"projectBuild":      map[string]any{keyType: typeString, keyEnum: catalog.ValidProjectBuilds},
								"groupId":           map[string]any{keyType: typeString},
								"artifactId":        map[string]any{keyType: typeString},
								"description":       map[string]any{keyType: typeString},
								"packageName":       map[string]any{keyType: typeString},
								"packaging":         map[string]any{keyType: typeString, keyEnum: catalog.ValidPackaging},
								"configuration":     map[string]any{keyType: typeString, keyEnum: catalog.SpringConfigFormats},
								"frameworkOptions": map[string]any{
									keyType:                 typeObject,
									keyAdditionalProperties: map[string]any{keyType: typeString},
									keyDescription:          "Key-value options specific to the service framework or runtime",
								},
								"dependencies": map[string]any{
									keyType:  typeArray,
									keyItems: map[string]any{keyType: typeString, keyEnum: catalog.ValidSpringDependencies},
								},
								"replicas": map[string]any{keyType: typeInteger, keyDefault: catalog.Defaults.Replicas},
								"resources": map[string]any{
									keyType:                 typeObject,
									keyAdditionalProperties: false,
									keyProperties: map[string]any{
										"cpu":    map[string]any{keyType: typeString, keyDefault: catalog.Defaults.CPU},
										"memory": map[string]any{keyType: typeString, keyDefault: catalog.Defaults.Memory},
									},
								},
								"uses": map[string]any{
									keyType:  typeArray,
									keyItems: map[string]any{keyType: typeString},
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
