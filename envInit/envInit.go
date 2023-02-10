package envInit

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func GetEnvString(key string) (string, error) {
	result := os.Getenv(key)
	if result == "" {
		return result, fmt.Errorf("Environment variable %s not set", key)
        }
	return result, nil
}

func SetEnvString(key string, value *string) error {
        result, err := GetEnvString(key)
        if err != nil {
            return err
        }
        *value = result
        return nil
}

func GetEnvInt(key string) (int, error) {
	envString, err := GetEnvString(key)
        if err != nil {
            return 0, err
        }
        return strconv.Atoi(envString)
}

func SetEnvInt(key string, value *int) error {
    result, err := GetEnvInt(key)
    if err != nil {
        return err
    }
    *value = result
    return nil
}

func GetEnvDuration(key string) (time.Duration, error) {
        envString, err := GetEnvString(key)
        if err != nil {
            return *new(time.Duration), err
        }
        return time.ParseDuration(envString)
}

func SetEnvDuration(key string, value *time.Duration) error {
        result, err := GetEnvDuration(key)
        if err != nil {
            return err
        }
        *value = result
        return nil
}

func GetEnvBool(key string) (bool, error) {
        envString, err := GetEnvString(key)
        if err != nil {
            return false, err
        }
        return strconv.ParseBool(envString)
}

func SetEnvBool(key string, value *bool) error {
    result, err := GetEnvBool(key)
    if err != nil {
        return err
    }
    *value = result
    return nil
}
