package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	RabbitMQ     RabbitMQConfig
	Stellar      StellarConfig
	Auth         AuthConfig
	Indexer      IndexerConfig
	Notification NotificationConfig
	CORS         CORSConfig
	RateLimit    RateLimitConfig
	Logging      LoggingConfig
	Environment  string
}

type ServerConfig struct {
	Port           int           `mapstructure:"port"`
	Host           string        `mapstructure:"host"`
	ReadTimeout    time.Duration `mapstructure:"readTimeout"`
	WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
	MaxHeaderBytes int           `mapstructure:"maxHeaderBytes"`
}

type DatabaseConfig struct {
	URL            string        `mapstructure:"url"`
	MaxOpenConns   int           `mapstructure:"maxOpenConns"`
	MaxIdleConns   int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
	MigrationPath  string        `mapstructure:"migrationPath"`
}

type RedisConfig struct {
	URL      string `mapstructure:"url"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"poolSize"`
}

type RabbitMQConfig struct {
	URL      string `mapstructure:"url"`
	Exchange string `mapstructure:"exchange"`
	Queues   struct {
		Notifications string `mapstructure:"notifications"`
		Webhooks      string `mapstructure:"webhooks"`
	} `mapstructure:"queues"`
}

type StellarConfig struct {
	Network           string `mapstructure:"network"`
	HorizonURL        string `mapstructure:"horizonUrl"`
	SorobanRPCURL     string `mapstructure:"sorobanRpcUrl"`
	NetworkPassphrase string `mapstructure:"networkPassphrase"`
	MasterPublicKey   string `mapstructure:"masterPublicKey"`
	MasterSecretKey   string `mapstructure:"masterSecretKey"`
}

type AuthConfig struct {
	JWTPrivateKeyPath string        `mapstructure:"jwtPrivateKeyPath"`
	JWTPublicKeyPath  string        `mapstructure:"jwtPublicKeyPath"`
	AccessTokenTTL    time.Duration `mapstructure:"accessTokenTTL"`
	RefreshTokenTTL   time.Duration `mapstructure:"refreshTokenTTL"`
	NonceTTL          time.Duration `mapstructure:"nonceTTL"`
}

type IndexerConfig struct {
	PollInterval time.Duration `mapstructure:"pollInterval"`
	BatchSize    int           `mapstructure:"batchSize"`
	StartLedger  int64         `mapstructure:"startLedger"`
}

type NotificationConfig struct {
	Email struct {
		Provider    string `mapstructure:"provider"`
		APIKey      string `mapstructure:"apiKey"`
		FromAddress string `mapstructure:"fromAddress"`
	} `mapstructure:"email"`
	SMS struct {
		Provider   string `mapstructure:"provider"`
		AccountSID string `mapstructure:"accountSid"`
		AuthToken  string `mapstructure:"authToken"`
		FromNumber string `mapstructure:"fromNumber"`
	} `mapstructure:"sms"`
	Push struct {
		FCMServerKey string `mapstructure:"fcmServerKey"`
	} `mapstructure:"push"`
}

type CORSConfig struct {
	AllowedOrigins   []string      `mapstructure:"allowedOrigins"`
	AllowedMethods   []string      `mapstructure:"allowedMethods"`
	AllowedHeaders   []string      `mapstructure:"allowedHeaders"`
	AllowCredentials bool          `mapstructure:"allowCredentials"`
	MaxAge           time.Duration `mapstructure:"maxAge"`
}

type RateLimitConfig struct {
	Global        int `mapstructure:"global"`
	Authenticated int `mapstructure:"authenticated"`
	Auth          int `mapstructure:"auth"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(path)
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/moistello/")
	v.SetEnvPrefix("MOISTELLO")
	v.AutomaticEnv()

	v.SetDefault("server.port", 1100)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.readTimeout", "10s")
	v.SetDefault("server.writeTimeout", "30s")
	v.SetDefault("server.maxHeaderBytes", 1048576)
	v.SetDefault("database.url", "postgres://moistello:moistello_dev@localhost:5432/moistello?sslmode=disable")
	v.SetDefault("database.maxOpenConns", 50)
	v.SetDefault("database.maxIdleConns", 10)
	v.SetDefault("database.connMaxLifetime", "30m")
	v.SetDefault("redis.url", "redis://localhost:6379")
	v.SetDefault("redis.poolSize", 20)
	v.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")
	v.SetDefault("rabbitmq.exchange", "moistello.events")
	v.SetDefault("rabbitmq.queues.notifications", "moistello.notifications")
	v.SetDefault("rabbitmq.queues.webhooks", "moistello.webhooks")
	v.SetDefault("stellar.network", "testnet")
	v.SetDefault("stellar.horizonUrl", "https://horizon-testnet.stellar.org")
	v.SetDefault("stellar.sorobanRpcUrl", "https://soroban-testnet.stellar.org")
	v.SetDefault("stellar.networkPassphrase", "Test SDF Network ; September 2015")
	v.SetDefault("auth.accessTokenTTL", "15m")
	v.SetDefault("auth.refreshTokenTTL", "168h")
	v.SetDefault("auth.nonceTTL", "5m")
	v.SetDefault("indexer.pollInterval", "3s")
	v.SetDefault("indexer.batchSize", 50)
	v.SetDefault("cors.allowedOrigins", []string{"http://localhost:1110"})
	v.SetDefault("cors.allowedMethods", []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowedHeaders", []string{"Authorization", "Content-Type", "X-Request-ID"})
	v.SetDefault("cors.allowCredentials", true)
	v.SetDefault("cors.maxAge", "24h")
	v.SetDefault("rateLimit.global", 100)
	v.SetDefault("rateLimit.authenticated", 300)
	v.SetDefault("rateLimit.auth", 10)
	v.SetDefault("logging.level", "debug")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("environment", "development")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	cfg.Environment = v.GetString("environment")
	return &cfg, nil
}
