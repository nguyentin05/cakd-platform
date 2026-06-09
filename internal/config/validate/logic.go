package validate

import (
	"fmt"
	"reflect"
)

// Logic performs Phase 3 of the configuration parsing pipeline.
// After the configuration struct has been structurally validated and populated with
// defaults, this phase enforces advanced business rules (e.g., cross-field dependencies)
// and relationship constraints (e.g., ensuring a service 'uses' an existing backing service).
func Logic(cfg any) error {
	if err := evaluateRules(cfg); err != nil {
		return err
	}

	v := reflect.ValueOf(cfg).Elem()

	backingField := v.FieldByName("Backing")
	backingNames := make(map[string]bool)
	for i := 0; i < backingField.Len(); i++ {
		backingNames[backingField.Index(i).FieldByName("Name").String()] = true
	}

	servicesField := v.FieldByName("Services")
	if servicesField.Len() == 0 {
		return fmt.Errorf("at least one service must be defined in services[]")
	}
	for i := 0; i < servicesField.Len(); i++ {
		svc := servicesField.Index(i)
		svcName := svc.FieldByName("Name").String()
		usesField := svc.FieldByName("Uses")
		for j := 0; j < usesField.Len(); j++ {
			use := usesField.Index(j).String()
			if !backingNames[use] {
				return fmt.Errorf("service %q references unknown backing resource %q", svcName, use)
			}
		}
	}

	return nil
}
