package utils

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

type EnvVariable int

const (
	ENV_ROOT EnvVariable = iota
	ENV_CONFIG_FILE
	ENV_MODULES_DIR
	ENV_SSH_KEY_PATH
	ENV_SSH_KEY_PASSWORD
)

var envMap = map[EnvVariable]string{
	ENV_ROOT:             "ENV_ROOT",
	ENV_CONFIG_FILE:      "CONFIG_FILE",
	ENV_MODULES_DIR:      "MODULES_DIR",
	ENV_SSH_KEY_PATH:     "SSH_KEY_PATH",
	ENV_SSH_KEY_PASSWORD: "SSH_KEY_PASSWORD",
}

func InitEnv() {
	envRoot := GetEnv(ENV_ROOT)

	err := godotenv.Load(filepath.Join(envRoot, "go.env.local"))
	if err != nil {
		log.Info(PrepareWarningOutput("FYI: couldn't open go.env.local"))
	}

	err = godotenv.Load(filepath.Join(envRoot, "go.env"))
	CheckError(err, "Error loading go.env file")
}

func GetEnv(env EnvVariable) string {
	return os.Getenv(envMap[env])
}
