package envInit

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type EnvKey string

func (key EnvKey)GetEnvString() (string, error) {
	result := os.Getenv(string(key))
	if result == "" {
		return result, fmt.Errorf("Environment variable %s not set", key)
	}
	return result, nil
}

func (key *EnvKey)GetEnvStringOrFallback(fallback string) (string, error) {
	result, err := key.GetEnvString()
	if err != nil {
		return fallback, err
	}
	return result, err
}

func (key *EnvKey)GetEnvInt() (int, error) {
	envString, err := key.GetEnvString()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(envString)
}

func (key *EnvKey)GetEnvIntOrFallback(fallback int) (int, error) {
	envString, err := key.GetEnvString()
	var result int
	result, err = strconv.Atoi(envString)
	if err != nil {
		return fallback, err
	}
	return result, err
}

func (key *EnvKey)GetEnvDuration() (time.Duration, error) {
	envString, err := key.GetEnvString()
	if err != nil {
		return *new(time.Duration), err
	}
	return time.ParseDuration(envString)
}

func (key *EnvKey)GetEnvDurationOrFallback(fallback time.Duration) (time.Duration, error) {
	envString, err := key.GetEnvString()
	result, err := time.ParseDuration(envString)
	if err != nil {
		return fallback, err
	}
	return result, err
}

func (key *EnvKey)GetEnvBool() (bool, error) {
	envString, err := key.GetEnvString()
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(envString)
}

func (key *EnvKey)GetEnvBoolOrFallback(fallback bool) (bool, error) {
	envString, err := key.GetEnvString()
	if err != nil {
		return fallback, err
	}
	return strconv.ParseBool(envString)
}
