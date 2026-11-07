package catalog_test

import (
	"slices"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
)

func TestCatalogContainsExpectedDefaultsAndProviders(t *testing.T) {
	if !slices.Contains(catalog.ClusterProviders, "aws-ec2") || !slices.Contains(catalog.ClusterProviders, "local") {
		t.Fatalf("expected aws-ec2 and local in ClusterProviders: %v", catalog.ClusterProviders)
	}
	if catalog.Defaults.LanguageVersion["java-spring-boot"] != "21" {
		t.Fatalf("expected default java version 21, got %q", catalog.Defaults.LanguageVersion["java-spring-boot"])
	}
	deps := catalog.ResolveSpringBootDependencies([]string{"web"}, []string{"postgresql"}, "prometheus")
	for _, want := range []string{"web", "data-jpa", "postgresql", "actuator", "prometheus"} {
		if !slices.Contains(deps, want) {
			t.Fatalf("expected dependency %q in resolved list %v", want, deps)
		}
	}
}
