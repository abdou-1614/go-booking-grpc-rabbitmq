package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	GRPCServer GRPCServer
	HttpServer HttpServer
	RabbitMQ   RabbitMQ
	Redis      RedisConfig
	Logger     Logger
	Postgres   PostgresConfig
	Jaeger     Jaeger
	Metrics    Metrics
}

type GRPCServer struct {
	AppVersion             string
	Port                   string
	CookieLifeTime         int
	AccessTokenExpire      int
	SessionID              string
	RefreshTokenExpire     int
	Mode                   string
	SessionPrefix          string
	CSRFPrefix             string
	Timeout                time.Duration
	ReadTimeout            time.Duration
	WriteTimeout           time.Duration
	MaxConnectionIdle      time.Duration
	MaxConnectionAge       time.Duration
	SessionGrpcServicePort string
}
type HttpServer struct {
	Port              string
	PprofPort         string
	Timeout           time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	CookieLifeTime    int
	SessionCookieName string
}

// RabbitMQ
type RabbitMQ struct {
	Host           string
	Port           string
	User           string
	Password       string
	Exchange       string
	Queue          string
	RoutingKey     string
	ConsumerTag    string
	WorkerPoolSize int
}

// Logger config
type Logger struct {
	Development       bool
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
}

// Postgresql config
type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  string
	PgDriver           string
}

// Redis config
type RedisConfig struct {
	RedisAddr      string
	RedisPassword  string
	RedisDB        string
	RedisDefaultDB string
	MinIdleConn    int
	PoolSize       int
	PoolTimeout    int
	Password       string
	DB             int
}

// Metrics config
type Metrics struct {
	URL         string
	ServiceName string
}

// Jaeger
type Jaeger struct {
	Host        string
	ServiceName string
	LogSpans    bool
}

func LoadConfig(fileName string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(fileName)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("config file not found")
		}
		return nil, err
	}

	return v, nil
}

func ParseConfig(v *viper.Viper) (*Config, error) {
	var c Config

	err := v.Unmarshal(&c)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func GetConfig(configFile string) (*Config, error) {
	cfgFile, err := LoadConfig(configFile)

	if err != nil {
		return nil, err
	}

	cfg, err := ParseConfig(cfgFile)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func GetConfigPath(configPath string) string {
	if configPath == "docker" {
		return "./config/config-docker"
	}

	return "./config/config-local"
}
