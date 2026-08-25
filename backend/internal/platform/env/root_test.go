package env

import (
	"testing"
)

type EnvConfig struct {
	Secret string `env:"SECRET"`
	Token  string `env:"TOKEN"`
}

func TestLoadParse(t *testing.T) {
	t.Setenv("SECRET", "nada")
	t.Setenv("TOKEN", "secret")

	envConfig := new(EnvConfig{})

	if err := LoadParse(envConfig); err != nil {
		t.Fatal(err)
	}
}
