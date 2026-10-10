package config

import "os"

func AppVersion() string {
	return os.Getenv("APP_VERSION")
}
