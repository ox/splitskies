package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type config struct {
	AppConfig    AppConfig    `yaml:"app"`
	DBConfig     DBConfig     `yaml:"db"`
	TwilioConfig TwilioConfig `yaml:"twilio"`
	// EmailConfig EmailConfig `yaml:"email"`
}

type AppConfig struct {
	LogLevel        string        `yaml:"log_level" env-default:"debug"`
	Bind            string        `yaml:"bind" env-default:"localhost:7077"`
	PublicPath      string        `yaml:"public" env-required:"true"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"5s"`
}

type EmailConfig struct {
	Host     string `yaml:"host" env-required:"true"`
	SMTPHost string `yaml:"smtp_host" env-required:"true"`
	From     string `yaml:"from" env-required:"true"`
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
}

type DBConfig struct {
	DBPath string `yaml:"path"`
}

type TwilioConfig struct {
	AccountSID        string `yaml:"account_sid"`
	AuthToken         string `yaml:"auth_token"`
	VerifyServicesSID string `yaml:"verify_services_sid"`
}

func FromFile(filepath string) (*config, error) {
	var cfg config

	err := cleanenv.ReadConfig(filepath, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
