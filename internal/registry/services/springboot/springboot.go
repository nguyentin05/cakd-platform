package springboot

import "github.com/nguyentin05/cakd-platform/internal/registry/backing"

// ConfigFormats defines valid configuration file formats supported for Spring Boot services.
var ConfigFormats = []string{"yaml", "properties"}

const dataJPA = "data-jpa"

// SpringBootDeps maps backing resource types to their Spring Boot starter dependencies.
var SpringBootDeps = map[string][]string{
	backing.PostgreSQL: {dataJPA, backing.PostgreSQL},
	backing.MySQL:      {dataJPA, backing.MySQL},
	backing.Redis:      {"data-redis"},
	backing.RabbitMQ:   {"amqp"},
}
