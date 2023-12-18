package utils

import "fmt"

type Env int64

const (
	EnvProd Env = iota
	EnvStaging
	EnvInvalid
)

func (env Env) String() string {
	switch env {
	case EnvProd:
		return "prod"
	case EnvStaging:
		return "staging"
	}
	return "unknown"
}

func EnvFromString(str string) (Env, error) {
	switch str {
	case "staging":
		{
			return EnvStaging, nil
		}
	case "prod":
		{
			return EnvProd, nil
		}
	}
	return EnvInvalid, fmt.Errorf("Invalid env: %s", str)
}
