package catalog

var (
	SpringConfigFormats  = []string{"yaml", "properties"}
	SpringBackingDepsMap = map[string][]string{
		BackingPostgreSQL: {depDataJPA, "postgresql"},
		BackingMySQL:      {depDataJPA, "mysql"},
		BackingRedis:      {"data-redis"},
		BackingRabbitMQ:   {"amqp"},
	}
	MonitoringDepsMap = map[string]string{
		ProviderPrometheus: "prometheus",
	}
)

// ResolveSpringBootDependencies merges user dependencies with backing and monitoring dependencies.
func ResolveSpringBootDependencies(userDeps, backingTypes []string, monitoringProvider string) []string {
	seen := make(map[string]bool, len(userDeps)+4)
	result := make([]string, 0, len(userDeps)+4)
	add := func(dep string) {
		if dep != "" && !seen[dep] {
			seen[dep] = true
			result = append(result, dep)
		}
	}

	for _, d := range userDeps {
		add(d)
	}
	for _, bt := range backingTypes {
		for _, dep := range SpringBackingDepsMap[bt] {
			add(dep)
		}
	}
	if monDep, ok := MonitoringDepsMap[monitoringProvider]; ok {
		add("actuator")
		add(monDep)
	}
	return result
}
