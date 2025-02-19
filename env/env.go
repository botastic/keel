package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Exists return true if env var is defined
func Exists(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

// MustExists panics if not exists
func MustExists(key string) {
	if _, ok := os.LookupEnv(key); !ok {
		panic(fmt.Sprintf("required environment variable %s does not exist", key))
	}
}

// Get env var or fallback
func Get(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// MustGet env var or panic
func MustGet(key string) string {
	MustExists(key)
	return Get(key, "")
}

// GetInt env var or fallback as int
func GetInt(key string, fallback int) int {
	if value, err := strconv.Atoi(Get(key, "")); err == nil {
		return value
	}
	return fallback
}

// MustGetInt env var as int or panic
func MustGetInt(key string) int {
	MustExists(key)
	return GetInt(key, 0)
}

// GetBool env var or fallback as bool
func GetBool(key string, fallback bool) bool {
	if val, err := strconv.ParseBool(Get(key, "")); err == nil {
		return val
	}
	return fallback
}

// MustGetBool env var as bool or panic
func MustGetBool(key string) bool {
	MustExists(key)
	return GetBool(key, false)
}

// GetStringSlice env var or fallback as []string
func GetStringSlice(key string, fallback []string) []string {
	if v := Get(key, ""); v != "" {
		return strings.Split(v, ",")
	}
	return fallback
}

// MustGetStringSlice env var as bool or panic
func MustGetStringSlice(key string) []string {
	MustExists(key)
	return GetStringSlice(key, nil)
}

// GetIntSlice env var or fallback as []string
func GetIntSlice(key string, fallback []int) []int {
	if v := Get(key, ""); v != "" {
		elements := strings.Split(v, ",")
		ret := make([]int, len(elements))
		for i, stringVal := range elements {
			intVal, err := strconv.Atoi(stringVal)
			if err != nil {
				return fallback
			}
			ret[i] = intVal
		}
	}
	return fallback
}

// MustGetGetIntSlice env var as bool or panic
func MustGetGetIntSlice(key string) []int {
	MustExists(key)
	return GetIntSlice(key, nil)
}
