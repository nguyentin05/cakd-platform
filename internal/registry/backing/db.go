package backing

const (
	PostgreSQL = "postgresql"
	MySQL      = "mysql"
	Redis      = "redis"
	RabbitMQ   = "rabbitmq"
)

// Types is the list of currently supported database types in the CAKD platform.
var Types = []string{
	PostgreSQL,
}
