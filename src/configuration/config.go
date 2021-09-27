package configuration

import (
    configParser "github.com/caarlos0/env/v6"
)


type Config struct {
    Environment string  `env:"ENV_NAME" envDefault:"development"`
    AccessToken string  `env:"ACCESS_TOKEN,notEmpty,unset"`
    Host string         `env:"HOST" envDefault:"localhost"`
    Port int16          `env:"PORT" envDefault:"3000"`
}


func New() (*Config, error) {
    cfg := Config{}

    if err := configParser.Parse( &cfg ); err != nil {
        return nil, err
    }

    return &cfg, nil
}
