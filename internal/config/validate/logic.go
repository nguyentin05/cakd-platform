package validate

import (
	"fmt"

	"github.com/nguyentin05/cakd-platform/internal/schema"
)

// Logic performs Phase 3 of the configuration parsing pipeline.
// After the configuration struct has been structurally validated and populated with
// defaults, this phase enforces advanced business rules (e.g., cross-field dependencies)
// and relationship constraints (e.g., ensuring a service 'uses' an existing backing service).
func Logic(cfg *schema.PlatformConfig) error {
	if err := evaluateRules(cfg); err != nil {
		return err
	}

	backingNames := make(map[string]bool)
	for _, b := range cfg.Backing {
		backingNames[b.Name] = true
	}

	if len(cfg.Services) == 0 {
		return fmt.Errorf("at least one service must be defined in services[]")
	}

	for _, svc := range cfg.Services {
		for _, use := range svc.Uses {
			if !backingNames[use] {
				return fmt.Errorf("service %q references unknown backing resource %q", svc.Name, use)
			}
		}
	}

	return nil
}
