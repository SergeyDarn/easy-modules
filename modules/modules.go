package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"easymodules/utils"

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
	log.Debug("Installing modules into " + getModulesDir())
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

	successMsg := "Installion of " + strconv.Itoa(len(modules)) + " modules took " + time.Since(start).String()
	log.Debug(utils.PrepareSuccessOutput(successMsg))
}

func installModule(moduleName string, moduleUrl string) {
	if !utils.IsGitUrl(moduleUrl) {
		return
	}

	sshKeyPath := utils.GetEnv(utils.ENV_SSH_KEY_PATH)
	sshKeyPasword := utils.GetEnv(utils.ENV_SSH_KEY_PASSWORD)
	moduleDir := filepath.Join(getModulesDir(), moduleName)
	_, err := os.Stat(moduleDir)
	isModuleNotCloned := os.IsNotExist(err)

	if isModuleNotCloned {
		utils.GitClone(moduleName, moduleUrl, moduleDir, sshKeyPath, sshKeyPasword)
		return
	}

	utils.CheckError(err, "Error while reading module "+moduleName+" folder")

	gitStatus := utils.GitDirStatus(moduleDir)
	if gitStatus.String() != "" {
		log.Info(
			utils.PrepareWarningOutput(
				"There are unsaved changes for module \"" + moduleName + "\" - skipping it. " +
					"UnsavedChanges =" + gitStatus.String(),
			),
		)
		return
	}

	err = os.RemoveAll(moduleDir)
	utils.CheckError(err, "Error while trying to delete module folder for "+moduleName)

	utils.GitClone(moduleName, moduleUrl, moduleDir, sshKeyPath, sshKeyPath)
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

	log.Info(
		utils.PrepareDangerOutput(
			"Unsaved Modules qty=" + strconv.Itoa(len(changedModules)) +
				"modules " + strings.Join(changedModules, " "),
		),
	)
}

func getModuleDir(moduleName string) string {
	return filepath.Join(getModulesDir(), moduleName)
}
