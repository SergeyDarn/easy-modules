package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestIsGitUrl(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"git", true},
		{"htts://google.com/", false},
		{"randomString", false},
		{"cool-js-module", false},
		{"https://github.com/SergeyDarn/scrape-search-ai.git", true},
		{"git@github.com:SergeyDarn/scrape-search-ai.git", true},
	}

	for _, test := range tests {
		testName := test.url

		t.Run(testName, func(t *testing.T) {
			res := IsGitUrl(test.url)

			if res != test.want {
				t.Errorf("Expected %t, but got %t", test.want, res)
			}
		})
	}
}

func TestParseGitUrl(t *testing.T) {
	type want struct {
		cleanUrl   string
		commitHash string
		branch     plumbing.ReferenceName
		tag        plumbing.ReferenceName
		error      bool
	}

	tests := []struct {
		name   string
		gitUrl string
		want   want
	}{
		{"Not a git url", "htts://google.com/", want{}},
		{"No reference http url", "https://github.com/SergeyDarn/scrape-search-ai.git", want{
			cleanUrl: "https://github.com/SergeyDarn/scrape-search-ai.git",
		}},
		{"No reference ssh url", "git@github.com:SergeyDarn/scrape-search-ai.git", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
		}},

		{"Branch 1", "git@github.com:SergeyDarn/scrape-search-ai.git#1", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			branch:   plumbing.NewBranchReferenceName("1"),
		}},
		{"Branch 123", "git@github.com:SergeyDarn/scrape-search-ai.git#123", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			branch:   plumbing.NewBranchReferenceName("123"),
		}},
		{"Branch dev", "git@github.com:SergeyDarn/scrape-search-ai.git#dev", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			branch:   plumbing.NewBranchReferenceName("dev"),
		}},

		{"Branch blablabla2", "git@github.com:SergeyDarn/scrape-search-ai.git#blablabla2", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			branch:   plumbing.NewBranchReferenceName("blablabla2"),
		}},
		{"Commit Hash", "git@github.com:SergeyDarn/scrape-search-ai.git#b7620f64a115b85eca08504cb9b364e594c9f8df", want{
			cleanUrl:   "git@github.com:SergeyDarn/scrape-search-ai.git",
			commitHash: "b7620f64a115b85eca08504cb9b364e594c9f8df",
		}},

		{"Tag 1.4.0", "git@github.com:SergeyDarn/scrape-search-ai.git#1.4.0", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			tag:      plumbing.NewTagReferenceName("1.4.0"),
		}},
		{"Tag 1.5", "git@github.com:SergeyDarn/scrape-search-ai.git#1.5", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			tag:      plumbing.NewTagReferenceName("1.5"),
		}},
		{"Tag 1.10.0.1", "git@github.com:SergeyDarn/scrape-search-ai.git#1.10.0.1", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			tag:      plumbing.NewTagReferenceName("1.10.0.1"),
		}},
		{"Tag 1.2.3_fix", "git@github.com:SergeyDarn/scrape-search-ai.git#1.2.3_fix", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			tag:      plumbing.NewTagReferenceName("1.2.3_fix"),
		}},

		{"Invalid Http url with # but no reference", "https://github.com/SergeyDarn/scrape-search-ai#", want{
			error: true,
		}},
		{"Invalid Ssh url with # but no reference", "git@github.com:SergeyDarn/scrape-search-ai.git#", want{
			error: true,
		}},
		{"Invalid tag 1.2.3#", "git@github.com:SergeyDarn/scrape-search-ai.git#1.2.3#", want{
			error: true,
		}},
		{"Invalid url with multiple #", "git@g#ithub.com:Serg#eyDarns#ear.git#11.3#", want{
			error: true,
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want.error {
				TestPanic(t, test.name, func() { parseGitUrl(test.gitUrl) })
				return
			}

			cleanUrl, commitHash, branch, tag := parseGitUrl(test.gitUrl)
			fail := false

			if cleanUrl != test.want.cleanUrl {
				t.Logf("Expected cleanUrl %s, but got %s", test.want.cleanUrl, cleanUrl)
				fail = true
			}

			if commitHash != test.want.commitHash {
				t.Logf("Expected commitHash %s, but got %s", test.want.commitHash, commitHash)
				fail = true
			}

			if branch != test.want.branch {
				t.Logf("Expected branch %s, but got %s", test.want.branch, branch)
				fail = true
			}

			if tag != test.want.tag {
				t.Logf("Expected tag %s, but got %s", test.want.tag, tag)
				fail = true
			}

			if fail {
				t.Fail()
			}
		})
	}
}

type gitCloneTest struct {
	name     string
	repoName string
	gitUrl   string
	want     gitCloneWant
}

type gitCloneWant struct {
	head   string
	tag    bool
	commit bool
	error  bool
}

func TestGitClone(t *testing.T) {
	testDir := "TEST_GIT_CLONE_DIR"
	tests := []gitCloneTest{
		{"Master/Main", "main", "https://github.com/SergeyDarn/test-module-js.git", gitCloneWant{
			head: plumbing.Main.Short(),
		}},
		{"SSH", "ssh", "git@github.com:SergeyDarn/test-module-js.git", gitCloneWant{
			head: plumbing.Main.Short(),
		}},
		{"Branch", "branch", "https://github.com/SergeyDarn/test-module-js.git#dev", gitCloneWant{
			head: "dev",
		}},
		{"Commit Hash", "commit", "https://github.com/SergeyDarn/test-module-js.git#5d62004178df760fd8978ef166e9ab14d23b06d1", gitCloneWant{
			head:   "5d62004178df760fd8978ef166e9ab14d23b06d1",
			commit: true,
		}},
		{"Tag", "tag", "https://github.com/SergeyDarn/test-module-js.git#1.0.0", gitCloneWant{
			head: "1.0.0",
			tag:  true,
		}},

		{"Not Git Url", "not_git", "^5.3.0", gitCloneWant{error: true}},
		{"Invalid Reference", "invalid_reference", "git@#github.#com:S#ergeyD#arn/test-module-js.git#", gitCloneWant{
			error: true,
		}},
		{"Not Existing Tag", "wrong_tag", "https://github.com/SergeyDarn/test-module-js.git#1.0.0_doesnt_exist", gitCloneWant{
			tag:   true,
			error: true,
		}},
		{"Not Existing Branch", "wrong_branch", "https://github.com/SergeyDarn/test-module-js.git#doesnt_exist", gitCloneWant{
			error: true,
		}},
		{"Not Existing Commit", "wrong_commit", "https://github.com/SergeyDarn/test-module-js.git#06a3e504bee6033d42690d01100edede077c7fe5", gitCloneWant{
			commit: true,
			error:  true,
		}},
	}

	t.Run("Git Clone Group", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				testGitClone(t, test, testDir)
			})
		}
	})

	os.RemoveAll(testDir)
}

func testGitClone(t *testing.T, test gitCloneTest, testDir string) {
	repoDir := filepath.Join(testDir, test.repoName)

	InitEnv()
	os.RemoveAll(repoDir)

	if test.want.error {
		TestPanic(t, test.name, func() {
			GitClone(test.repoName, test.gitUrl, repoDir)
		})

		return
	}

	GitClone(test.repoName, test.gitUrl, repoDir)

	_, err := os.Stat(repoDir)
	CheckTestError(t, err)

	repo, err := git.PlainOpen(repoDir)
	CheckTestError(t, err)

	headName := GetHeadShort(repo, test.want.commit, test.want.tag)

	if headName != test.want.head {
		t.Errorf("Expected HEAD to be %s, but got %s", test.want.head, headName)
	}
}
