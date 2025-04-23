package main

import (
	"flag"

	"easymodules/modules"
	"easymodules/utils"

	"github.com/charmbracelet/log"
)

func main() {
	log.SetLevel(log.DebugLevel)

	utils.InitEnv()

	showChangedModules := *flag.Bool("show-changed-modules", false, "run command to show modules with unsaved git changes")
	parallelInstall := *flag.Bool("parallel-install", true, "install modules in parallel (true/false)")
	safeInstall := *flag.Bool("safe-install", true, "if this is set to false, modules folder will be deleted on start. Default version - each module is checked separately, and only if module has no unsaved changes, it's deleted and then reinstalled")
	flag.Parse()

	if showChangedModules {
		modules.ShowChangedModules()
		return
	}

	configJson := modules.ReadConfigJson()
	var dependencies map[string]string = utils.MergeMaps(
		configJson.Dependencies,
		configJson.DevDependencies,
	)

	if !safeInstall {
		modules.RemoveModulesDir()
	}

	modules.CreateModulesDir()
	modules.InstallModules(dependencies, parallelInstall)
}
