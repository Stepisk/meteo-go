package config

import (
	"github.com/spf13/viper"
	"time"
)

const (
	defaultHttpPort               = "8000"
	defaultHttpReadWriteTimeout   = 10 * time.Second
	defaultHttpMaxHeaderMegabytes = 1

	EnvLocal = "local"
)

type (
	Config struct {
		Environment string
	}
)

// Init populates Config struct with values from config file
// located at filepath and environment variables
func Init(configsDir string) (*Config, error) {
	populateDefaults()

	if err := parseEnv(); err != nil {
		return nil, err
	}

	if err := parseConfigFile(configsDir, viper.GetString("env")); err != nil {
		return nil, err
	}

	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	setFromEnv(&cfg)

	return &cfg, nil
}

func unmarshal(cfg *Config) error {
	return nil
}

func setFromEnv(cfg *Config) {
	cfg.Environment = viper.GetString("env")
}

func parseConfigFile(folder, env string) error {
	viper.AddConfigPath(folder)
	viper.SetConfigName("main")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if env == EnvLocal {
		return nil
	}

	viper.SetConfigName(env)
	return viper.MergeInConfig()
}

func populateDefaults() {
	viper.SetDefault("http.port", defaultHttpPort)
	viper.SetDefault("http.max_header_megabytes", defaultHttpMaxHeaderMegabytes)
	viper.SetDefault("http.timeouts.read", defaultHttpReadWriteTimeout)
	viper.SetDefault("http.timeouts.write", defaultHttpReadWriteTimeout)
}

func parseEnv() error {
	if err := parseAppEnvFromEnv(); err != nil {
		return err
	}

	if err := parseStorageEnvVariables(); err != nil {
		return err
	}

	return parsePasswordFromEnv()
}

func parsePasswordFromEnv() error {
	viper.SetEnvPrefix("password")
	return viper.BindEnv("salt")
}

func parseAppEnvFromEnv() error {
	viper.SetEnvPrefix("app")
	return viper.BindEnv("env")
}

func parseStorageEnvVariables() error {
	viper.SetEnvPrefix("storage")
	if err := viper.BindEnv("bucket"); err != nil {
		return err
	}

	if err := viper.BindEnv("endpoint"); err != nil {
		return err
	}

	return viper.BindEnv("secret_key")
}
