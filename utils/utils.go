package utils

import (
	"os"

	"github.com/charmbracelet/log"
)

func CheckError(err error, additionalMessage string) {
	if err == nil {
		return
	}

	if additionalMessage != "" {
		log.Error(additionalMessage + ": " + err.Error())
	}

	os.Exit(1)
}

func MergeMaps(map1 map[string]string, map2 map[string]string) map[string]string {
	for name, value := range map2 {
		map1[name] = value
	}

	return map1
}
