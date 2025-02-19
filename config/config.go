package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// config holds the global configuration
var config *viper.Viper

// Init sets up the configuration
func init() {
	config = viper.New()
	config.AutomaticEnv()
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// Config return the config instance
func Config() *viper.Viper {
	return config
}

func GetBool(c *viper.Viper, key string, fallback bool) func() bool {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() bool {
		return c.GetBool(key)
	}
}

func MustGetBool(c *viper.Viper, key string, fallback bool) func() bool {
	c = ensure(c)
	must(c, key)
	return func() bool {
		return c.GetBool(key)
	}
}

func GetInt(c *viper.Viper, key string, fallback int) func() int {
	c.SetDefault(key, fallback)
	return func() int {
		return c.GetInt(key)
	}
}

func MustGetInt(c *viper.Viper, key string) func() int {
	must(c, key)
	return func() int {
		return c.GetInt(key)
	}
}

func GetString(c *viper.Viper, key, fallback string) func() string {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() string {
		return c.GetString(key)
	}
}

func MustGetString(c *viper.Viper, key string) func() string {
	c = ensure(c)
	must(c, key)
	return func() string {
		return c.GetString(key)
	}
}

func GetStringSlice(c *viper.Viper, key string, fallback []string) func() []string {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() []string {
		return c.GetStringSlice(key)
	}
}

func ensure(c *viper.Viper) *viper.Viper {
	if c == nil {
		c = config
	}
	return c
}

func must(c *viper.Viper, key string) {
	if !c.IsSet(key) {
		panic(fmt.Sprintf("missing required config key: %s", key))
	}
}
