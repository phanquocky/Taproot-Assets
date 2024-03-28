package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type (
	Config struct {
		Env string `env-required:"true" env:"ENV"`
		Network
		Mongo
	}

	Network struct {
		WalletPassphrase string `env:"WALLET_PASSPHRASE"`
		UserConfig       string `env:"USER_CONFIG"`
		PassConfig       string `env:"PASS_CONFIG"`
		SenderAddrSim    string `env:"SENDER_ADDR_SIM_CONFIG"`
		SenderAddrTest   string `env:"SENDER_ADDR_TEST_CONFIG"`
	}

	Mongo struct {
		ConnURI string `env-required:"true" env:"MONGO_CONN_URI"`
		DBName  string `env-required:"true" env:"MONGO_DB_NAME"`
	}
)

func NewConfig() *Config {
	cfg := &Config{}
	if err := godotenv.Load(); err != nil {
		cfg = &Config{}
	}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		panic(err)
		return nil
	}

	return cfg
}
