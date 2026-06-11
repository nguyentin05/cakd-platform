package registry

// SpringBootDeps maps backing resource types to their Spring Boot starter dependencies.
// This is the Single Source of Truth — no hardcoded string comparisons elsewhere.
var SpringBootDeps = map[string][]string{
	PostgreSQL: {"data-jpa", PostgreSQL},
	MySQL:      {"data-jpa", MySQL},
	Redis:      {"data-redis"},
	RabbitMQ:   {"amqp"},
}

// MonitoringDeps maps monitoring providers to their Spring Boot starter dependencies.
var MonitoringDeps = map[string]string{
	Prometheus: Prometheus,
}
