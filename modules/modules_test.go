package modules

import (
	"easymodules/utils"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/log"
)

type installModuleTest struct {
	name       string
	moduleName string
	moduleUrl  string
	want       installModuleWant
}

type installModuleWant struct {
	noDir         bool
	testGitStatus bool
}

func TestInstallModule(t *testing.T) {
	tests := []installModuleTest{
		{"Not a git module", "not_a_git_module", "5.0.0", installModuleWant{
			noDir: true,
		}},
		{"Simple install", "simple_install", "https://github.com/SergeyDarn/test-module-js.git", installModuleWant{}},
		{"Git status", "git_status", "https://github.com/SergeyDarn/test-module-js.git", installModuleWant{
			testGitStatus: true,
		}},
	}

	utils.InitEnv()
	log.SetLevel(log.ErrorLevel)

	rootModulesDir := utils.GetPathRootDir(getModulesDir())
	os.RemoveAll(rootModulesDir)

	t.Run("InstallModule Group", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				testInstallModule(t, test)
			})
		}
	})

	os.RemoveAll(rootModulesDir)
}

func testInstallModule(t *testing.T, test installModuleTest) {
	moduleDir := getModuleDir(test.moduleName)

	var initialGitStatus string
	if test.want.testGitStatus {
		installModuleWithChanges(t, test, moduleDir)

		initialGitStatus = utils.GitDirStatus(moduleDir).String()
	}

	installModule(test.moduleName, test.moduleUrl)

	err, _ := checkModuleDirStatus(moduleDir)
	if !test.want.noDir && err != nil {
		t.Fatalf("Expected module to exist after install, but got error: %s", err.Error())
	}

	if test.want.noDir || !test.want.testGitStatus {
		return
	}

	gitStatus := utils.GitDirStatus(moduleDir).String()

	if (initialGitStatus != "") && (gitStatus != initialGitStatus) {
		t.Fatalf("Expected module %s that has git changes to preserve them after installModule call", test.moduleName)
	}
}

func installModuleWithChanges(t *testing.T, test installModuleTest, moduleDir string) {
	testFile := "test.txt"

	installModule(test.moduleName, test.moduleUrl)

	err, _ := checkModuleDirStatus(moduleDir)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = os.Create(filepath.Join(moduleDir, testFile))
	if err != nil {
		t.Fatal(err.Error())
	}
}
