package utils

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

type EnvVariable int

const (
	ENV_CONFIG_FILE EnvVariable = iota
	ENV_MODULES_DIR
	ENV_SSH_KEY_PATH
	ENV_SSH_KEY_PASSWORD
)

var envMap = map[EnvVariable]string{
	ENV_CONFIG_FILE:      "CONFIG_FILE",
	ENV_MODULES_DIR:      "MODULES_DIR",
	ENV_SSH_KEY_PATH:     "SSH_KEY_PATH",
	ENV_SSH_KEY_PASSWORD: "SSH_KEY_PASSWORD",
}

func InitEnv() {
	err := godotenv.Load("go.env.local")
	if err != nil {
		log.Info("FYI: couldn't open go.env.local")
	}

	err = godotenv.Load("go.env")
	CheckError(err, "Error loading go.env file")
}

func GetEnv(env EnvVariable) string {
	return os.Getenv(envMap[env])
}
