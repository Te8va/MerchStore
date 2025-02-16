package config

type Config struct {
	ServerAddress    string `env:"SERVER_ADDRESS" envDefault:"0.0.0.0:8080"`
	ServicePort      int    `env:"SERVICE_PORT"          envDefault:"8080"`
	ServiceHost      string `env:"SERVICE_HOST"          envDefault:"0.0.0.0"`
	PostgresUser     string `env:"POSTGRES_USER"     envDefault:"merch"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"     envDefault:"merch"`
	PostgresDB       string `env:"POSTGRES_DATABASE"           envDefault:"merch"`
	PostgresPort     int    `env:"POSTGRES_PORT"         envDefault:"5432"`
	MigrationsPath   string `env:"MIGRATIONS_PATH"       envDefault:"migrations"`
	LogFilePath      string `env:"LOG_FILE_PATH"         envDefault:"logfile.log"`
	JWTKey           string `env:"JWT_KEY"               envDefault:"supermegasecret"`
	PostgresConn     string `env:"POSTGRES_CONN" envDefault:"postgres://merch:merch@merch-db:5432/merch?sslmode=disable"`
}
