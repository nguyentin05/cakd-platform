package registry

// Condition specifies the operator type evaluated by the rule engine (e.g. NotEmpty, IsTrue).
type Condition string

const (
	NotEmpty Condition = "NotEmpty"
	IsTrue   Condition = "IsTrue"
	NotNil   Condition = "NotNil"
)

// DependencyRule defines a conditional relationship between two configuration paths
// which must be satisfied to pass business logic validation.
type DependencyRule struct {
	IfPath   string    // Path to the condition field (e.g., "Providers.CI")
	IfCond   Condition // The condition that triggers the rule
	ThenPath string    // Path to the dependent field (e.g., "Providers.CD")
	ThenCond Condition // The condition that must be met
	ErrorMsg string    // Error message if the rule is violated
}

// BusinessRules defines the list of cross-field constraints evaluated during validation.
var BusinessRules = []DependencyRule{
	{
		IfPath:   "Providers.CD",
		IfCond:   NotEmpty,
		ThenPath: "Providers.CI",
		ThenCond: NotEmpty,
		ErrorMsg: "providers.ci is required when providers.cd is set",
	},
	{
		IfPath:   "Observability.Alerting",
		IfCond:   IsTrue,
		ThenPath: "Providers.Notification",
		ThenCond: NotEmpty,
		ErrorMsg: "providers.notification is required when observability.alerting is enabled",
	},
	{
		IfPath:   "Observability.AI",
		IfCond:   NotNil,
		ThenPath: "Providers.LLM",
		ThenCond: NotEmpty,
		ErrorMsg: "providers.llm is required when observability.ai is enabled",
	},
	{
		IfPath:   "Observability.AI",
		IfCond:   NotNil,
		ThenPath: "Observability.AI.Model",
		ThenCond: NotEmpty,
		ErrorMsg: "observability.ai.model is required when observability.ai is enabled",
	},
}
