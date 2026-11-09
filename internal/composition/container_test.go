package composition_test

import (
	"context"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/composition"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

type mockVCS struct {
	pushed bool
	rolled bool
}

func (m *mockVCS) ProvisionAndPush(_ context.Context, _ ports.RepoProvisionRequest) (string, error) {
	m.pushed = true
	return "mock://pushed", nil
}

func (m *mockVCS) Rollback(_ context.Context, _ ports.RepoProvisionRequest) error {
	m.rolled = true
	return nil
}

func TestNewContainerAndToServices(t *testing.T) {
	c := composition.NewContainer()
	if c == nil {
		t.Fatal("expected NewContainer to return non-nil container")
	}

	if c.CatalogUC == nil {
		t.Fatal("expected CatalogUC to be initialized in container")
	}

	s := c.ToServices()
	if s == nil {
		t.Fatal("expected ToServices to return non-nil services bundle")
	}
	if s.CatalogUC != c.CatalogUC {
		t.Fatal("expected CatalogUC to match container")
	}
	if s.CreateUC != c.CreateUC {
		t.Fatal("expected CreateUC to match container")
	}
}

func TestVCSRouter(t *testing.T) {
	const provLocalGit = "local-git"
	ghMock := &mockVCS{}
	localMock := &mockVCS{}

	router := composition.NewVCSRouter(map[string]ports.VCSProvider{
		"github":     ghMock,
		provLocalGit: localMock,
	})

	// Test routing to local-git
	url, err := router.ProvisionAndPush(context.Background(), ports.RepoProvisionRequest{
		Provider: provLocalGit,
	})
	if err != nil || !localMock.pushed {
		t.Fatalf("expected local-git to be called, got err: %v, url: %s", err, url)
	}

	// Test routing to default github when provider empty
	url, err = router.ProvisionAndPush(context.Background(), ports.RepoProvisionRequest{})
	if err != nil || !ghMock.pushed {
		t.Fatalf("expected default github to be called, got err: %v, url: %s", err, url)
	}

	// Test rollback
	_ = router.Rollback(context.Background(), ports.RepoProvisionRequest{Provider: provLocalGit})
	if !localMock.rolled {
		t.Fatalf("expected local rollback to be called")
	}
}
