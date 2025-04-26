package modules

import (
	"easymodules/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

type JsonConfig struct {
	Dependencies    map[string]string
	DevDependencies map[string]string
}

var MODULES_DIR_PERMISSIONS os.FileMode = 0o777

func getModulesDir() string {
	return utils.GetEnv(utils.ENV_MODULES_DIR)
}

func ReadConfigJson() JsonConfig {
	configJson, fileErr := os.ReadFile(utils.GetEnv(utils.ENV_CONFIG_FILE))
	utils.CheckError(fileErr, "Error when opening json configuration file")

	var configJsonParsed JsonConfig
	jsonErr := json.Unmarshal(configJson, &configJsonParsed)
	utils.CheckError(jsonErr, "Error when parsing json configuration file")

	return configJsonParsed
}

func CreateModulesDir() {
	err := os.MkdirAll(getModulesDir(), MODULES_DIR_PERMISSIONS)
	utils.CheckError(err, "Error when creating modules folder")
}

func RemoveModulesDir() {
	err := os.RemoveAll(getModulesDir())
	utils.CheckError(err, "Error while trying to delete modules folder")
	log.Debug(utils.PrepareDangerOutput("Modules folder deleted before installation"))
}

func InstallModules(modules map[string]string, parallelInstall bool) {
	log.Debugf("Installing modules into %s", getModulesDir())
	fmt.Println()

	start := time.Now()

	if !parallelInstall {
		for name, url := range modules {
			installModule(name, url)
		}
	} else {
		var waitGroup sync.WaitGroup

		for name, url := range modules {
			waitGroup.Add(1)

			go func() {
				installModule(name, url)
				defer waitGroup.Done()
			}()
		}

		waitGroup.Wait()
	}

	log.Debugf(
		utils.PrepareSuccessOutput("Installation of %d modules took %s"),
		len(modules),
		time.Since(start),
	)
}

func installModule(moduleName string, moduleUrl string) {
	if !utils.IsGitUrl(moduleUrl) {
		return
	}

	moduleDir := filepath.Join(getModulesDir(), moduleName)
	_, err := os.Stat(moduleDir)
	isModuleNotCloned := os.IsNotExist(err)

	if isModuleNotCloned {
		utils.GitClone(moduleName, moduleUrl, moduleDir)
		return
	}

	utils.CheckError(err, "Error while reading module "+moduleName+" folder")

	gitStatus := utils.GitDirStatus(moduleDir)
	if gitStatus.String() != "" {
		log.Infof(
			utils.PrepareWarningOutput(
				"\nThere are unsaved changes for module \"%s\" - skipping it\n"+
					"\n%s",
			),
			moduleName,
			gitStatus.String(),
		)
		return
	}

	err = os.RemoveAll(moduleDir)
	utils.CheckError(err, "Error while trying to delete module folder for "+moduleName)

	utils.GitClone(moduleName, moduleUrl, moduleDir)
}

func ShowChangedModules() {
	modules, err := os.ReadDir(getModulesDir())
	utils.CheckError(err, "Error reading modules folder")
	changedModules := []string{}

	for _, module := range modules {
		gitStatus := utils.GitDirStatus(
			getModuleDir(module.Name()),
		)

		if gitStatus.String() != "" {
			changedModules = append(changedModules, module.Name())
		}
	}

	if len(changedModules) == 0 {
		log.Info(utils.PrepareSuccessOutput("You have no unsaved modules"))
		return
	}

	log.Infof(
		utils.PrepareWarningOutput("\nUnsaved Modules (%d):\n\n%s"),
		len(changedModules),
		strings.Join(changedModules, "\n"),
	)
}

func getModuleDir(moduleName string) string {
	return filepath.Join(getModulesDir(), moduleName)
}
